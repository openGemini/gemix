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
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joomcode/errorx"
	"github.com/openGemini/gemix/pkg/cluster/config"
	"github.com/openGemini/gemix/pkg/cluster/ctxt"
	"github.com/openGemini/gemix/pkg/cluster/operation"
	"github.com/openGemini/gemix/pkg/cluster/spec"
	"github.com/openGemini/gemix/pkg/cluster/task"
	"github.com/openGemini/gemix/utils"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

// StartCluster start the cluster with specified name.
func (m *Manager) StartCluster(name string, gOpt operation.Options, fn ...func(b *task.Builder, metadata spec.Metadata)) error {
	m.logger.Info("Starting cluster ...", zap.String("cluster name", name))

	// check locked
	//if err := m.specManager.ScaleOutLockedErr(name); err != nil {
	//	return err
	//}

	metadata, err := m.meta(name)
	if err != nil {
		return err
	}

	topo := metadata.GetTopology()
	base := metadata.GetBaseMeta()

	//tlsCfg, err := topo.TLSConfig(m.specManager.Path(name, spec.TLSCertKeyDir))
	//if err != nil {
	//	return err
	//}

	b, err := m.sshTaskBuilder(name, topo, base.User, gOpt)
	if err != nil {
		return err
	}

	b.Func("StartCluster", func(ctx context.Context) error {
		return operation.Start(ctx, topo, gOpt, nil)
	})

	for _, f := range fn {
		f(b, metadata)
	}

	t := b.Build()

	ctx := ctxt.New(
		context.Background(),
		gOpt.Concurrency,
		m.logger,
	)
	if err := t.Execute(ctx); err != nil {
		if errorx.Cast(err) != nil {
			// FIXME: Map possible task errors and give suggestions.
			return err
		}
		return errors.WithStack(err)
	}

	m.logger.Info("Started cluster successfully", zap.String("cluster name", name))
	return nil
}

type Starter interface {
	PrepareForStart() error
	Start() error
	Close()
}

type GeminiStarter struct {
	version string
	// ip -> remotes
	remotes map[string]*config.RemoteHost
	runs    *operation.RunActions

	// ip -> ssh clients
	sshClients map[string]*ssh.Client

	configurator config.Configurator // conf reader
	executor     operation.Executor  // execute commands on remote host

	clusterOptions utils.ClusterOptions
	startOptions   utils.StartOptions

	wg sync.WaitGroup
}

func NewGeminiStarter(ops utils.ClusterOptions, startOpts utils.StartOptions) Starter {
	return &GeminiStarter{
		remotes:        make(map[string]*config.RemoteHost),
		sshClients:     make(map[string]*ssh.Client),
		version:        ops.Version,
		configurator:   config.NewGeminiConfigurator(ops.YamlPath, filepath.Join(utils.DownloadDst, ops.Version, utils.LocalEtcRelPath, utils.LocalConfName), filepath.Join(utils.DownloadDst, ops.Version, utils.LocalEtcRelPath)),
		runs:           &operation.RunActions{},
		clusterOptions: ops,
		startOptions:   startOpts,
	}
}

func (d *GeminiStarter) PrepareForStart() error {
	var err error
	if err = d.configurator.BuildConfig(); err != nil {
		return err
	}
	conf := d.configurator.GetConfig()

	d.startOptions.User = conf.CommonConfig.User

	if err = d.prepareRemotes(conf, false); err != nil {
		fmt.Printf("Failed to establish SSH connections with all remote servers. The specific error is: %s\n", err)
		return err
	}
	fmt.Println("Success to establish SSH connections with all remote servers.")

	d.executor = operation.NewGeminiExecutor(d.sshClients)

	// check process conflict and port conflict
	if d.checkProcessConflict() {
		return errors.New("process conflict before starting")
	}
	if conflicted, port, ip := d.checkPortConflict(conf); conflicted {
		return fmt.Errorf("port %d conflict in %s before starting", port, ip)
	}

	if err = d.prepareRunActions(conf); err != nil {
		return err
	}

	// create user if needed
	if err = d.createUserIfNeed(); err != nil {
		return err
	}

	return nil
}

func (d *GeminiStarter) createUserIfNeed() error {
	if !d.startOptions.SkipCreateUser {
		newUserName := d.startOptions.User
		createUserCommand := fmt.Sprintf("useradd %s -s /sbin/nologin", newUserName)
		for ip := range d.remotes {
			_, err := d.executor.ExecCommand(ip, createUserCommand)
			if err != nil {
				fmt.Printf("Failed to create a new user on %s, error: %s", ip, err)
				return err
			}
		}
	}
	return nil
}

func (d *GeminiStarter) prepareRemotes(c *config.Config, needSftp bool) error {
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

func (d *GeminiStarter) tryConnect() error {
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

func (d *GeminiStarter) prepareRunActions(c *config.Config) error {
	// ts-meta
	i := 1
	for _, host := range c.CommonConfig.MetaHosts {
		d.runs.MetaAction = append(d.runs.MetaAction, &operation.RunAction{
			User: d.startOptions.User,
			Info: &operation.RunInfo{
				ScriptPath: filepath.Join(d.remotes[host].UpDataPath, d.version, utils.RemoteEtcRelPath, utils.InstallScript),
				Args: []string{utils.TsMeta, d.remotes[host].LogPath,
					filepath.Join(d.remotes[host].UpDataPath, d.version, utils.RemoteBinRelPath, utils.TsMeta),
					filepath.Join(d.remotes[host].UpDataPath, d.version, utils.RemoteEtcRelPath, utils.RemoteMetaConfName),
					filepath.Join(d.remotes[host].LogPath, utils.RemotePidPath, utils.TsMeta+utils.RemotePidSuffix),
					filepath.Join(d.remotes[host].LogPath, utils.MetaExtraLog+utils.RemoteLogSuffix)},
			},
			Remote: d.remotes[host],
		})
		i++
	}

	// ts-sql
	i = 1
	for _, host := range c.CommonConfig.SqlHosts {
		d.runs.SqlAction = append(d.runs.SqlAction, &operation.RunAction{
			User: d.startOptions.User,
			Info: &operation.RunInfo{
				ScriptPath: filepath.Join(d.remotes[host].UpDataPath, d.version, utils.RemoteEtcRelPath, utils.InstallScript),
				Args: []string{utils.TsSql, d.remotes[host].LogPath,
					filepath.Join(d.remotes[host].UpDataPath, d.version, utils.RemoteBinRelPath, utils.TsSql),
					filepath.Join(d.remotes[host].UpDataPath, d.version, utils.RemoteEtcRelPath, utils.RemoteSqlConfName),
					filepath.Join(d.remotes[host].LogPath, utils.RemotePidPath, utils.TsSql+utils.RemotePidSuffix),
					filepath.Join(d.remotes[host].LogPath, utils.SqlExtraLog+utils.RemoteLogSuffix)},
			},
			Remote: d.remotes[host],
		})
		i++
	}

	// ts-store
	i = 1
	for _, host := range c.CommonConfig.StoreHosts {
		d.runs.StoreAction = append(d.runs.StoreAction, &operation.RunAction{
			User: d.startOptions.User,
			Info: &operation.RunInfo{
				ScriptPath: filepath.Join(d.remotes[host].UpDataPath, d.version, utils.RemoteEtcRelPath, utils.InstallScript),
				Args: []string{utils.TsStore, d.remotes[host].LogPath,
					filepath.Join(d.remotes[host].UpDataPath, d.version, utils.RemoteBinRelPath, utils.TsStore),
					filepath.Join(d.remotes[host].UpDataPath, d.version, utils.RemoteEtcRelPath, utils.RemoteStoreConfName),
					filepath.Join(d.remotes[host].LogPath, utils.RemotePidPath, utils.TsStore+utils.RemotePidSuffix),
					filepath.Join(d.remotes[host].LogPath, utils.StoreExtraLog+utils.RemoteLogSuffix)},
			},
			Remote: d.remotes[host],
		})
		i++
	}

	return nil
}

func (d *GeminiStarter) checkProcessConflict() bool {
	for ip := range d.remotes {
		output, err := d.executor.ExecCommand(ip, CheckProcessCommand)
		if err != nil {
			fmt.Println(err)
			return true
		} else {
			if output != "" {
				fmt.Printf("process conflict in remote %s\n", ip)
				fmt.Println(output)
				return true
			}
		}
	}
	return false
}

func (d *GeminiStarter) checkPortConflict(conf *config.Config) (bool, int, string) {
	// check port conflict about ts-meta
	for _, ip := range conf.CommonConfig.MetaHosts {
		tomlPath := filepath.Join(utils.DownloadDst, d.version, utils.LocalEtcRelPath, ip, utils.RemoteMetaConfName)
		t, err := config.ReadFromToml(tomlPath)
		if err != nil {
			fmt.Println(err)
			return true, 0, ip
		}

		if conflicted, port, err := d.checkOnePortWithStr(ip, t.Meta.BindAddress); err != nil {
			fmt.Println(err)
			return true, 0, ip
		} else {
			if conflicted {
				return true, port, ip
			}
		}

		if conflicted, port, err := d.checkOnePortWithStr(ip, t.Meta.HttpBindAddress); err != nil {
			fmt.Println(err)
			return true, 0, ip
		} else {
			if conflicted {
				return true, port, ip
			}
		}

		if conflicted, port, err := d.checkOnePortWithStr(ip, t.Meta.RpcBindAddress); err != nil {
			fmt.Println(err)
			return true, 0, ip
		} else {
			if conflicted {
				return true, port, ip
			}
		}

		if conflicted, port, err := d.checkOnePortWithInt(ip, t.Gossip.MetaBindPort); err != nil {
			fmt.Println(err)
			return true, 0, ip
		} else {
			if conflicted {
				return true, port, ip
			}
		}
	}

	// check port conflict about ts-sql
	for _, ip := range conf.CommonConfig.SqlHosts {
		tomlPath := filepath.Join(utils.DownloadDst, d.version, utils.LocalEtcRelPath, ip, utils.RemoteSqlConfName)
		t, err := config.ReadFromToml(tomlPath)
		if err != nil {
			fmt.Println(err)
			return true, 0, ip
		}

		if conflicted, port, err := d.checkOnePortWithStr(ip, t.Http.BindAddress); err != nil {
			fmt.Println(err)
			return true, 0, ip
		} else {
			if conflicted {
				return true, port, ip
			}
		}
	}

	// check port conflict about ts-store
	for _, ip := range conf.CommonConfig.StoreHosts {
		tomlPath := filepath.Join(utils.DownloadDst, d.version, utils.LocalEtcRelPath, ip, utils.RemoteStoreConfName)
		t, err := config.ReadFromToml(tomlPath)
		if err != nil {
			fmt.Println(err)
			return true, 0, ip
		}

		if conflicted, port, err := d.checkOnePortWithStr(ip, t.Data.StoreIngestAddr); err != nil {
			fmt.Println(err)
			return true, 0, ip
		} else {
			if conflicted {
				return true, port, ip
			}
		}

		if conflicted, port, err := d.checkOnePortWithStr(ip, t.Data.StoreSelectAddr); err != nil {
			fmt.Println(err)
			return true, 0, ip
		} else {
			if conflicted {
				return true, port, ip
			}
		}

		if conflicted, port, err := d.checkOnePortWithInt(ip, t.Gossip.StoreBindPort); err != nil {
			fmt.Println(err)
			return true, 0, ip
		} else {
			if conflicted {
				return true, port, ip
			}
		}
	}

	return false, 0, ""
}

func (d *GeminiStarter) checkOnePortWithStr(ip, inputStr string) (bool, int, error) {
	parts := strings.Split(inputStr, ":")
	if len(parts) < 2 {
		return true, 0, errors.New("invalid inputStr format when check one port conflict")
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
	return true, port, errors.New("unexpected port conflict check result")
}

func (d *GeminiStarter) checkOnePortWithInt(ip string, port int) (bool, int, error) {
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
	return true, port, errors.New("unexpected port conflict check result")
}

func (d *GeminiStarter) Start() error {
	if d.executor == nil {
		return nil
	}
	errChan := make(chan error, len(d.runs.MetaAction)+len(d.runs.SqlAction)+len(d.runs.StoreAction))
	var wgp sync.WaitGroup
	wgp.Add(2)

	go func() {
		defer wgp.Done()
		// start all ts-meta concurrently
		d.wg.Add(len(d.runs.MetaAction))
		for _, action := range d.runs.MetaAction {
			go func(action *operation.RunAction, errChan chan error) {
				defer d.wg.Done()
				d.executor.ExecRunAction(action, errChan)
			}(action, errChan)
		}
		d.wg.Wait()

		// time for ts-meta campaign
		time.Sleep(time.Second)

		// start all ts-store and ts-sql concurrently
		d.wg.Add(len(d.runs.SqlAction) + len(d.runs.StoreAction))
		for _, action := range d.runs.StoreAction {
			go func(action *operation.RunAction, errChan chan error) {
				defer d.wg.Done()
				d.executor.ExecRunAction(action, errChan)
			}(action, errChan)
		}
		for _, action := range d.runs.SqlAction {
			go func(action *operation.RunAction, errChan chan error) {
				defer d.wg.Done()
				d.executor.ExecRunAction(action, errChan)
			}(action, errChan)
		}
		d.wg.Wait()
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
		return errors.New("start cluster failed")
	} else {
		return nil
	}
}

func (d *GeminiStarter) Close() {
	var err error
	for _, ssh := range d.sshClients {
		if err = ssh.Close(); err != nil {
			fmt.Println(err)
		}
	}
}
