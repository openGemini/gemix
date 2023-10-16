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

package operation

import (
	"fmt"

	"github.com/openGemini/gemix/pkg/cluster/config"
	"github.com/openGemini/gemix/util"
	"golang.org/x/crypto/ssh"
)

type Executor interface {
	ExecRunAction(action *RunAction, errChan chan error) string
	ExecStopAction(action *StopAction) (string, error)
	ExecCommand(ip string, command string) (string, error)
}

type GeminiExecutor struct {
	sshClients map[string]*ssh.Client
}

func NewGeminiExecutor(clients map[string]*ssh.Client) Executor {
	new := &GeminiExecutor{
		sshClients: clients,
	}
	return new
}

func (e *GeminiExecutor) ExecCommand(ip string, command string) (string, error) {
	sshClient := e.sshClients[ip]
	if sshClient == nil {
		fmt.Printf("no ssh client for %s\n", ip)
		return "", util.ErrNoSshClient
	}

	sshSession, err := util.NewSshSession(sshClient)
	if err != nil {
		fmt.Printf("new ssh session failed! %v", err)
	}
	defer sshSession.Close()

	combo, err := sshSession.CombinedOutput(command)
	if err != nil {
		fmt.Printf("exec: %s on %s failed! %v\n", command, ip, err)
		return "", err
	}
	return string(combo), nil
}

type RunInfo struct {
	ScriptPath string
	Args       []string
}

type RunAction struct {
	Info   *RunInfo
	Remote *config.RemoteHost
}

type RunActions struct {
	MetaAction  []*RunAction
	SqlAction   []*RunAction
	StoreAction []*RunAction
}

func (e *GeminiExecutor) ExecRunAction(action *RunAction, errChan chan error) string {
	ip := action.Remote.Ip
	sshClient := e.sshClients[ip]
	if sshClient == nil {
		fmt.Printf("no ssh client for %s\n", ip)
		errChan <- util.ErrNoSshClient
		return ""
	}

	sshSession, err := util.NewSshSession(sshClient)
	if err != nil {
		fmt.Printf("new ssh session failed! %v", err)
	}
	defer sshSession.Close()

	// generate command
	command := fmt.Sprintf("chmod a+x %s; /bin/bash %s", action.Info.Args[2], action.Info.ScriptPath)
	for _, arg := range action.Info.Args {
		command = fmt.Sprintf("%s %s", command, arg)
	}

	combo, err := sshSession.CombinedOutput(command)
	if err != nil {
		fmt.Printf("exec: %s on %s failed! %v\n", command, ip, err)
		errChan <- err
		return ""
	}
	return string(combo)
}

type StopAction struct {
	ProcessNames []string
	Remote       *config.RemoteHost
}

func (e *GeminiExecutor) ExecStopAction(action *StopAction) (string, error) {
	ip := action.Remote.Ip
	sshClient := e.sshClients[ip]
	if sshClient == nil {
		fmt.Printf("no ssh client for %s\n", ip)
		return "", util.ErrNoSshClient
	}

	command := ""
	for _, bin := range action.ProcessNames {
		command = fmt.Sprintf(`%s pgrep -x %s > /dev/null && killall %s && echo "Terminating process: %s" || echo "Process %s not found.";`, command, bin, bin, bin, bin)
	}

	sshSession, err := util.NewSshSession(sshClient)
	if err != nil {
		fmt.Printf("new ssh session failed! %v", err)
	}
	defer sshSession.Close()
	combo, err := sshSession.CombinedOutput(command)
	if err != nil {
		fmt.Printf("exec: %s on %s failed! %v\n", command, ip, err)
		return "", err
	}

	return string(combo), nil
}
