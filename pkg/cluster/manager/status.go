// Copyright 2023 Huawei Cloud Computing Technologies Co., Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package manager

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/olekukonko/tablewriter"
	"github.com/openGemini/gemix/pkg/cluster/config"
	"github.com/openGemini/gemix/pkg/cluster/operation"
	"github.com/openGemini/gemix/utils"
	"golang.org/x/crypto/ssh"
)

const CheckProcessCommand = "ps aux | grep -E '(ts-meta|ts-sql|ts-store)' | grep -v grep | awk '{print $11}'"
const CheckDiskCapacityCommand = "df -h | grep '^/dev/'"

func GenCheckPortCommand(port int) string {
	return fmt.Sprintf("ss -tln | grep -q ':%d' && echo 'yes' || echo 'no'", port)
}

type ClusterStatusPerServer struct {
	Ip               string
	RunningProcesses []string     // ts-meta,ts-sql,ts-store
	PortOccupancy    map[int]bool // port->occupancy or not
	DiskCapacity     []string     // disk->capacity
}

type StatusPatroller interface {
	PrepareForPatrol() error
	Patrol() error
	Close()
}

type GeminiStatusPatroller struct {
	version string
	conf    *config.Config
	// ip -> remotes
	remotes map[string]*config.RemoteHost

	// ip -> ssh clients
	sshClients map[string]*ssh.Client

	configurator config.Configurator // conf reader
	executor     operation.Executor  // execute commands on remote host

	clusterOptions utils.ClusterOptions

	wg sync.WaitGroup
}

func NewGeminiStatusPatroller(ops utils.ClusterOptions) StatusPatroller {
	return &GeminiStatusPatroller{
		version:        ops.Version,
		remotes:        make(map[string]*config.RemoteHost),
		sshClients:     make(map[string]*ssh.Client),
		configurator:   config.NewGeminiConfigurator(ops.YamlPath, "", ""),
		clusterOptions: ops,
	}
}

func (d *GeminiStatusPatroller) PrepareForPatrol() error {
	var err error
	if err = d.configurator.BuildConfig(); err != nil {
		return err
	}
	d.conf = d.configurator.GetConfig()

	// check the internet with all the remote servers
	if err = d.prepareRemotes(d.conf); err != nil {
		fmt.Printf("Failed to establish SSH connections with all remote servers. The specific error is: %s\n", err)
		return err
	}
	fmt.Println("Success to establish SSH connections with all remote servers.")

	d.executor = operation.NewGeminiExecutor(d.sshClients)
	return nil
}

func (d *GeminiStatusPatroller) prepareRemotes(c *config.Config) error {
	if c == nil {
		return utils.ErrUnexpectedNil
	}

	for ip, ssh := range c.SSHConfig {
		d.remotes[ip] = &config.RemoteHost{
			Ip:         ip,
			SSHPort:    ssh.Port,
			UpDataPath: ssh.UpDataPath,
			LogPath:    ssh.LogPath,
			User:       d.clusterOptions.User,
			Typ:        d.clusterOptions.SshType,
			Password:   d.clusterOptions.Password,
			KeyPath:    d.clusterOptions.Key,
		}
	}

	if err := d.tryConnect(); err != nil {
		return err
	}

	return nil
}

func (d *GeminiStatusPatroller) tryConnect() error {
	for ip, r := range d.remotes {
		var err error
		var sshClient *ssh.Client
		switch r.Typ {
		case utils.SSH_PW:
			sshClient, err = utils.NewSSH_PW(r.User, r.Password, r.Ip, r.SSHPort)
		case utils.SSH_KEY:
			sshClient, err = utils.NewSSH_Key(r.User, r.KeyPath, r.Ip, r.SSHPort)

		}
		if err != nil {
			return err
		}
		d.sshClients[ip] = sshClient
	}
	return nil
}

func (d *GeminiStatusPatroller) Patrol() error {
	statusChan := make(chan ClusterStatusPerServer, len(d.remotes))
	errChan := make(chan error, len(d.remotes))
	var wgp sync.WaitGroup
	wgp.Add(3)
	go func() {
		defer wgp.Done()
		d.wg.Add(len(d.remotes))
		for ip := range d.remotes {
			go d.patrolOneServer(ip, statusChan, errChan)
		}
		d.wg.Wait()
		close(errChan)
		close(statusChan)
	}()

	var has_err = false
	go func() {
		defer wgp.Done()
		for {
			err, ok := <-errChan
			if !ok {
				return
			}
			fmt.Println(err)
			has_err = true
		}
	}()

	go func() {
		defer wgp.Done()
		for {
			status, ok := <-statusChan
			if !ok {
				return
			}
			displayGeminiStatus(status)
		}
	}()

	wgp.Wait()
	if has_err {
		return errors.New("check cluster status failed")
	}
	return nil
}

func (d *GeminiStatusPatroller) patrolOneServer(ip string, statusChan chan ClusterStatusPerServer, errChan chan error) {
	defer d.wg.Done()
	var status = ClusterStatusPerServer{
		Ip:            ip,
		PortOccupancy: make(map[int]bool),
	}
	var err error

	// check running process
	output, err := d.executor.ExecCommand(ip, CheckProcessCommand)
	if err != nil {
		fmt.Println(err)
		errChan <- err
		return
	} else {
		status.RunningProcesses = strings.Split(output, "\n")
	}

	// check port status about ts-meta
	for _, i := range d.conf.CommonConfig.MetaHosts {
		if ip == i {
			tomlPath := filepath.Join(utils.DownloadDst, d.version, utils.LocalEtcRelPath, ip, utils.RemoteMetaConfName)
			t, err := config.ReadFromToml(tomlPath)
			if err != nil {
				fmt.Println(err)
				errChan <- err
				return
			}

			if occupied, port, err := d.checkOnePortWithStr(ip, t.Meta.BindAddress); err != nil {
				fmt.Println(err)
				errChan <- err
				return
			} else {
				status.PortOccupancy[port] = occupied
			}

			if occupied, port, err := d.checkOnePortWithStr(ip, t.Meta.HttpBindAddress); err != nil {
				fmt.Println(err)
				errChan <- err
				return
			} else {
				status.PortOccupancy[port] = occupied
			}

			if occupied, port, err := d.checkOnePortWithStr(ip, t.Meta.RpcBindAddress); err != nil {
				fmt.Println(err)
				errChan <- err
				return
			} else {
				status.PortOccupancy[port] = occupied
			}

			if occupied, port, err := d.checkOnePortWithInt(ip, t.Gossip.MetaBindPort); err != nil {
				fmt.Println(err)
				errChan <- err
				return
			} else {
				status.PortOccupancy[port] = occupied
			}
		}
	}
	// check port status about ts-sql
	for _, i := range d.conf.CommonConfig.SqlHosts {
		if ip == i {
			tomlPath := filepath.Join(utils.DownloadDst, d.version, utils.LocalEtcRelPath, ip, utils.RemoteSqlConfName)
			t, err := config.ReadFromToml(tomlPath)
			if err != nil {
				fmt.Println(err)
				errChan <- err
				return
			}

			if occupied, port, err := d.checkOnePortWithStr(ip, t.Http.BindAddress); err != nil {
				fmt.Println(err)
				errChan <- err
				return
			} else {
				status.PortOccupancy[port] = occupied
			}
		}
	}
	// check port status about ts-store
	for _, i := range d.conf.CommonConfig.StoreHosts {
		if ip == i {
			tomlPath := filepath.Join(utils.DownloadDst, d.version, utils.LocalEtcRelPath, ip, utils.RemoteStoreConfName)
			t, err := config.ReadFromToml(tomlPath)
			if err != nil {
				fmt.Println(err)
				errChan <- err
				return
			}

			if occupied, port, err := d.checkOnePortWithStr(ip, t.Data.StoreIngestAddr); err != nil {
				fmt.Println(err)
				errChan <- err
				return
			} else {
				status.PortOccupancy[port] = occupied
			}

			if occupied, port, err := d.checkOnePortWithStr(ip, t.Data.StoreSelectAddr); err != nil {
				fmt.Println(err)
				errChan <- err
				return
			} else {
				status.PortOccupancy[port] = occupied
			}

			if occupied, port, err := d.checkOnePortWithInt(ip, t.Gossip.StoreBindPort); err != nil {
				fmt.Println(err)
				errChan <- err
				return
			} else {
				status.PortOccupancy[port] = occupied
			}
		}
	}

	// check disk capacity
	output, err = d.executor.ExecCommand(ip, CheckDiskCapacityCommand)
	if err != nil {
		fmt.Println(err)
		errChan <- err
		return
	} else {
		status.DiskCapacity = strings.Split(output, "\n")
	}
	statusChan <- status
}

func (d *GeminiStatusPatroller) checkOnePortWithStr(ip, inputStr string) (bool, int, error) {
	parts := strings.Split(inputStr, ":")
	if len(parts) < 2 {
		return true, 0, errors.New("invalid inputStr format when check one port status")
	}
	portStr := parts[1]

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return true, port, err
	}

	output, err := d.executor.ExecCommand(ip, GenCheckPortCommand(port))
	if err != nil {
		return true, port, err
	} else {
		if strings.Contains(output, "yes") {
			return true, port, nil
		} else if strings.Contains(output, "no") {
			return false, port, nil
		}
	}
	return true, port, errors.New("unexpected port status result")
}

func (d *GeminiStatusPatroller) checkOnePortWithInt(ip string, port int) (bool, int, error) {
	output, err := d.executor.ExecCommand(ip, GenCheckPortCommand(port))
	if err != nil {
		return true, port, err
	} else {
		if strings.Contains(output, "yes") {
			return true, port, nil
		} else if strings.Contains(output, "no") {
			return false, port, nil
		}
	}
	return true, port, errors.New("unexpected port status result")
}

func (d *GeminiStatusPatroller) Close() {
	var err error
	for _, ssh := range d.sshClients {
		if err = ssh.Close(); err != nil {
			fmt.Println(err)
		}
	}
}

func displayGeminiStatus(status ClusterStatusPerServer) {
	fmt.Printf("\nGemini status of server %s\n", status.Ip)

	// Create a new table for Running Processes
	runningProcessesTable := tablewriter.NewWriter(os.Stdout)
	runningProcessesTable.SetHeader([]string{"Running Processes"})
	for _, process := range status.RunningProcesses {
		runningProcessesTable.Append([]string{fmt.Sprintf("%v", process)})
	}
	runningProcessesTable.Render()

	// Create a new table for Port Occupancy
	portOccupancyTable := tablewriter.NewWriter(os.Stdout)
	portOccupancyTable.SetHeader([]string{"Port Occupancy"})
	for port, occupied := range status.PortOccupancy {
		portOccupancyTable.Append([]string{fmt.Sprintf("Port %d: Occupied: %v", port, occupied)})
	}
	portOccupancyTable.Render()

	// Create a new table for Disk Capacity
	diskCapacityTable := tablewriter.NewWriter(os.Stdout)
	diskCapacityTable.SetHeader([]string{"Disk Capacity"})
	for _, capacity := range status.DiskCapacity {
		diskCapacityTable.Append([]string{fmt.Sprintf("%v", capacity)})
	}
	diskCapacityTable.Render()
}
