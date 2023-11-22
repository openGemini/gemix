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
	"fmt"
	"sync"

	"github.com/openGemini/gemix/pkg/cluster/config"
	"github.com/openGemini/gemix/pkg/cluster/operation"
	"github.com/openGemini/gemix/utils"
	"golang.org/x/crypto/ssh"
)

type Stop interface {
	Prepare() error
	Run() error
	Close()
}

type GeminiStop struct {
	// ip -> remotes
	remotes map[string]*config.RemoteHost

	// ip -> actions
	stops map[string]*operation.StopAction

	// ip -> ssh clients
	sshClients map[string]*ssh.Client

	configurator config.Configurator // conf reader
	executor     operation.Executor  // execute commands on remote host

	wg sync.WaitGroup

	clusterOptions utils.ClusterOptions
}

func NewGeminiStop(ops utils.ClusterOptions) Stop {
	new := &GeminiStop{
		remotes:        make(map[string]*config.RemoteHost),
		stops:          make(map[string]*operation.StopAction),
		sshClients:     make(map[string]*ssh.Client),
		configurator:   config.NewGeminiConfigurator(ops.YamlPath, "", ""),
		clusterOptions: ops,
	}
	return new
}

func (s *GeminiStop) Prepare() error {
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

	if err = s.prepareStopActions(conf); err != nil {
		return err
	}

	return nil
}

func (s *GeminiStop) prepareRemotes(c *config.Config) error {
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
	}

	if err := s.tryConnect(); err != nil {
		return err
	}

	return nil
}

func (s *GeminiStop) tryConnect() error {
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

func (s *GeminiStop) prepareStopActions(c *config.Config) error {

	// ts-meta
	for ip := range c.SSHConfig {
		if s.stops[ip] == nil {
			s.stops[ip] = &operation.StopAction{
				Remote: s.remotes[ip],
			}
		}
		s.stops[ip].ProcessNames = append(s.stops[ip].ProcessNames, utils.TsMeta)
	}

	// ts-sql
	for ip := range c.SSHConfig {
		if s.stops[ip] == nil {
			s.stops[ip] = &operation.StopAction{
				Remote: s.remotes[ip],
			}
		}
		s.stops[ip].ProcessNames = append(s.stops[ip].ProcessNames, utils.TsSql)
	}

	// ts-store
	for ip := range c.SSHConfig {
		if s.stops[ip] == nil {
			s.stops[ip] = &operation.StopAction{
				Remote: s.remotes[ip],
			}
		}
		s.stops[ip].ProcessNames = append(s.stops[ip].ProcessNames, utils.TsStore)
	}
	return nil
}

func (s *GeminiStop) Run() error {
	if s.executor == nil {
		return utils.ErrUnexpectedNil
	}
	s.wg.Add(len(s.stops))
	for _, action := range s.stops {
		go func(action *operation.StopAction) {
			defer s.wg.Done()
			s.executor.ExecStopAction(action)
		}(action)
	}
	s.wg.Wait()
	return nil
}

func (s *GeminiStop) Close() {
	var err error
	for _, ssh := range s.sshClients {
		if ssh != nil {
			if err = ssh.Close(); err != nil {
				fmt.Println(err)
			}
		}
	}
}
