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
	"strings"

	"github.com/fatih/color"
	operator "github.com/openGemini/gemix/pkg/cluster/operation"
	"github.com/openGemini/gemix/pkg/cluster/spec"
	"github.com/openGemini/gemix/pkg/cluster/task"
	"github.com/openGemini/gemix/pkg/gui"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Manager to deploy a cluster.
type Manager struct {
	sysName     string
	specManager *spec.SpecManager
	//bindVersion spec.BindVersion
	logger *zap.Logger
}

// NewManager create a Manager.
func NewManager(sysName string, specManager *spec.SpecManager, logger *zap.Logger) *Manager {
	return &Manager{
		sysName:     sysName,
		specManager: specManager,
		logger:      logger,
	}
}

func (m *Manager) meta(name string) (metadata spec.Metadata, err error) {
	exist, err := m.specManager.Exist(name)
	if err != nil {
		return nil, err
	}

	if !exist {
		return nil, errors.Errorf("%s cluster `%s` not exists", m.sysName, name)
	}

	metadata = m.specManager.NewMetadata() // TODO: 没有可用信息
	err = m.specManager.Metadata(name, metadata)
	if err != nil {
		return metadata, err
	}

	return metadata, nil
}

func (m *Manager) confirmTopology(clusterName, version string, topo spec.Topology) error {
	fmt.Println("Please confirm your topology:")

	cyan := color.New(color.FgCyan, color.Bold)
	//fmt.Printf("Cluster type:    %s\n", cyan.Sprint(m.sysName))
	fmt.Printf("Cluster name:    %s\n", cyan.Sprint(clusterName))
	fmt.Printf("Cluster version: %s\n", cyan.Sprint(version))

	clusterTable := [][]string{
		// Header
		{"Role", "Host"},
	}
	clusterTable[0] = append(clusterTable[0], "Ports", "OS/Arch", "Directories")

	topo.IterInstance(func(instance spec.Instance) {
		comp := instance.ComponentName()
		instInfo := []string{comp, instance.GetHost()}
		instInfo = append(instInfo,
			strings.Trim(strings.Replace(fmt.Sprint(instance.UsedPorts()), " ", "/", -1), "[]"),
			fmt.Sprintf("%s/%s", instance.OS(), instance.Arch()),
			strings.Join(instance.UsedDirs(), ","))

		clusterTable = append(clusterTable, instInfo)
	})

	gui.PrintTable(clusterTable, true)

	fmt.Println("Attention:")
	fmt.Println("    1. If the topology is not what you expected, check your yaml file.")
	fmt.Println("    2. Please confirm there is no port/directory conflicts in same host.")

	return gui.PromptForConfirmOrAbortError("Do you want to continue? [y/N]: ")
}

func (m *Manager) sshTaskBuilder(name string, topo spec.Topology, user string, gOpt operator.Options) (*task.Builder, error) {
	//var p *gui.SSHConnectionProps = &gui.SSHConnectionProps{}
	//if gOpt.SSHType != executor.SSHTypeNone && len(gOpt.SSHProxyHost) != 0 {
	//	var err error
	//	if p, err = gui.ReadIdentityFileOrPassword(gOpt.SSHProxyIdentity, gOpt.SSHProxyUsePassword); err != nil {
	//		return nil, err
	//	}
	//}

	return task.NewBuilder(m.logger).
		SSHKeySet(
			m.specManager.Path(name, "ssh", "id_rsa"),
			m.specManager.Path(name, "ssh", "id_rsa.pub"),
		).
		ClusterSSH(
			topo,
			user,
			gOpt.SSHTimeout,
			gOpt.OptTimeout,
		), nil
}

// fillHost full host cpu-arch and kernel-name
func (m *Manager) fillHost(s *gui.SSHConnectionProps, topo spec.Topology, user string) error {
	hostArchOrOS := map[string]string{}
	topo.IterInstance(func(instance spec.Instance) {
		insOS := instance.OS()
		if insOS == "" {
			insOS = topo.BaseTopo().GlobalOptions.OS
		}
		hostArchOrOS[instance.GetHost()] = insOS
	})
	if err := topo.FillHostArchOrOS(hostArchOrOS, spec.FullOSType); err != nil {
		return errors.WithStack(err)
	}

	hostArchOrOS = map[string]string{}
	topo.IterInstance(func(instance spec.Instance) {
		insArch := instance.Arch()
		if insArch == "" {
			insArch = topo.BaseTopo().GlobalOptions.Arch
		}
		hostArchOrOS[instance.GetHost()] = insArch
	})
	if err := topo.FillHostArchOrOS(hostArchOrOS, spec.FullArchType); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// fillHostArchOrOS full host cpu-arch or kernel-name
//func (m *Manager) fillHostArchOrOS(s, p *tui.SSHConnectionProps, topo spec.Topology, gOpt *operator.Options, user string, fullType spec.FullHostType) error {
//	globalSSHType := topo.BaseTopo().GlobalOptions.SSHType
//	hostArchOrOS := map[string]string{}
//	var detectTasks []*task.StepDisplay
//
//	topo.IterInstance(func(inst spec.Instance) {
//		if fullType == spec.FullOSType {
//			if inst.OS() != "" {
//				return
//			}
//		} else if inst.Arch() != "" {
//			return
//		}
//
//		if _, ok := hostArchOrOS[inst.GetManageHost()]; ok {
//			return
//		}
//		hostArchOrOS[inst.GetManageHost()] = ""
//
//		tf := task.NewBuilder(m.logger).
//			RootSSH(
//				inst.GetManageHost(),
//				inst.GetSSHPort(),
//				user,
//				s.Password,
//				s.IdentityFile,
//				s.IdentityFilePassphrase,
//				gOpt.SSHTimeout,
//				gOpt.OptTimeout,
//				gOpt.SSHProxyHost,
//				gOpt.SSHProxyPort,
//				gOpt.SSHProxyUser,
//				p.Password,
//				p.IdentityFile,
//				p.IdentityFilePassphrase,
//				gOpt.SSHProxyTimeout,
//				gOpt.SSHType,
//				globalSSHType,
//			)
//
//		switch fullType {
//		case spec.FullOSType:
//			tf = tf.Shell(inst.GetManageHost(), "uname -s", "", false)
//		default:
//			tf = tf.Shell(inst.GetManageHost(), "uname -m", "", false)
//		}
//		detectTasks = append(detectTasks, tf.BuildAsStep(fmt.Sprintf("  - Detecting node %s %s info", inst.GetManageHost(), string(fullType))))
//	})
//	if len(detectTasks) == 0 {
//		return nil
//	}
//
//	ctx := ctxt.New(
//		context.Background(),
//		gOpt.Concurrency,
//		m.logger,
//	)
//	t := task.NewBuilder(m.logger).
//		ParallelStep(fmt.Sprintf("+ Detect CPU %s Name", string(fullType)), false, detectTasks...).
//		Build()
//
//	if err := t.Execute(ctx); err != nil {
//		return perrs.Annotate(err, "failed to fetch cpu-arch or kernel-name")
//	}
//
//	for host := range hostArchOrOS {
//		stdout, _, ok := ctxt.GetInner(ctx).GetOutputs(host)
//		if !ok {
//			return fmt.Errorf("no check results found for %s", host)
//		}
//		hostArchOrOS[host] = strings.Trim(string(stdout), "\n")
//	}
//	return topo.FillHostArchOrOS(hostArchOrOS, fullType)
//}
