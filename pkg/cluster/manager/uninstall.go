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
	"sync"

	"github.com/openGemini/gemix/pkg/cluster/config"
	"github.com/openGemini/gemix/pkg/cluster/operation"
	"github.com/openGemini/gemix/utils"
	"golang.org/x/crypto/ssh"
)

type Uninstall interface {
	Prepare() error
	Run() error
	Close()
}

type GeminiUninstaller struct {
	// ip -> remotes
	remotes map[string]*config.RemoteHost
	// ip -> ssh clients
	sshClients map[string]*ssh.Client

	configurator config.Configurator // conf reader
	executor     operation.Executor  // execute commands on remote host
	upDataPath   map[string]string   // ip->up path

	wg sync.WaitGroup

	clusterOptions utils.ClusterOptions
}

func NewGeminiUninstaller(ops utils.ClusterOptions) Uninstall {
	new := &GeminiUninstaller{
		remotes:        make(map[string]*config.RemoteHost),
		sshClients:     make(map[string]*ssh.Client),
		configurator:   config.NewGeminiConfigurator(ops.YamlPath, "", ""),
		upDataPath:     make(map[string]string),
		clusterOptions: ops,
	}
	return new
}

func (s *GeminiUninstaller) Prepare() error {
	var err error
	if err = s.configurator.BuildConfig(); err != nil {
		return err
	}
	conf := s.configurator.GetConfig()

	if err = s.prepareRemotes(conf); err != nil {
		fmt.Printf("Failed to establish SSH connections with all remote servers. The specific error is: %s\n", err)
		return err
	}
	fmt.Println("Success to establish SSH connections with all remote servers.")

	s.executor = operation.NewGeminiExecutor(s.sshClients)

	return nil
}

func (s *GeminiUninstaller) prepareRemotes(c *config.Config) error {
	if c == nil {
		return utils.ErrUnexpectedNil
	}

	for ip, ssh := range c.SSHConfig {
		s.remotes[ip] = &config.RemoteHost{
			Ip:       ip,
			SSHPort:  ssh.Port,
			User:     s.clusterOptions.User,
			Password: s.clusterOptions.Password,
			KeyPath:  s.clusterOptions.Key,
			Typ:      s.clusterOptions.SshType,
		}

		s.upDataPath[ip] = ssh.UpDataPath
	}

	if err := s.tryConnect(); err != nil {
		return err
	}

	return nil
}

func (s *GeminiUninstaller) tryConnect() error {
	for ip, r := range s.remotes {
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
		s.sshClients[ip] = sshClient
	}
	return nil
}

func (s *GeminiUninstaller) Run() error {
	if s.executor == nil {
		return utils.ErrUnexpectedNil
	}

	errChan := make(chan error, len(s.remotes))
	var wgp sync.WaitGroup
	wgp.Add(2)

	go func() {
		defer wgp.Done()
		s.wg.Add(len(s.remotes))
		for ip := range s.remotes {
			go func(ip string, errChan chan error) {
				defer s.wg.Done()
				filePath := filepath.Join(s.upDataPath[ip], s.clusterOptions.Version)
				if filePath == "/" || filePath == "/root" {
					errChan <- fmt.Errorf("can not remove %s on %s", filePath, ip)
					return
				}
				command := fmt.Sprintf("rm -rf %s;", filePath)
				_, err := s.executor.ExecCommand(ip, command)
				if err != nil {
					errChan <- err
				}
			}(ip, errChan)
		}
		s.wg.Wait()
		close(errChan)
	}()

	var has_err = false
	go func() {
		defer wgp.Done()
		for {
			err, ok := <-errChan
			if !ok {
				break
			}
			fmt.Println(err)
			has_err = true
		}
	}()

	wgp.Wait()
	if has_err {
		return errors.New("uninstall cluster failed")
	} else {
		err := os.Remove(filepath.Join(utils.ClusterInfoDir, s.clusterOptions.Name))
		if err != nil {
			return errors.New("error deleting the cluster info file")
		}
		return nil
	}
}

func (s *GeminiUninstaller) Close() {
	var err error
	for _, ssh := range s.sshClients {
		if ssh != nil {
			if err = ssh.Close(); err != nil {
				fmt.Println(err)
			}
		}
	}
}
