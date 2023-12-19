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
	"path/filepath"
	"strings"

	operator "github.com/openGemini/gemix/pkg/cluster/operation"
	"github.com/openGemini/gemix/pkg/cluster/spec"
	"github.com/openGemini/gemix/pkg/cluster/task"
	"github.com/openGemini/gemix/pkg/gui"
	logprinter "github.com/openGemini/gemix/pkg/logger/printer"
	"github.com/openGemini/gemix/pkg/meta"
	"github.com/openGemini/gemix/pkg/set"
)

// buildEnvInitTasks builds the EnvInit tasks
func buildEnvInitTasks(topo spec.Topology, opt *InstallOptions, gOpt *operator.Options, sshConnProps *gui.SSHConnectionProps, logger *logprinter.Logger) []*task.StepDisplay {
	base := topo.BaseTopo()
	globalOptions := base.GlobalOptions

	uniqueHosts := getAllUniqueHosts(topo)

	var envInitTasks []*task.StepDisplay

	for host, info := range uniqueHosts {
		var dirs []string
		for _, dir := range []string{globalOptions.DeployDir, globalOptions.LogDir, globalOptions.DataDir} {
			if dir == "" {
				continue
			}

			dirs = append(dirs, spec.Abs(globalOptions.User, dir))
		}

		t := task.NewBuilder(logger).
			RootSSH(
				host,
				info.ssh,
				opt.User,
				sshConnProps.Password,
				sshConnProps.IdentityFile,
				sshConnProps.IdentityFilePassphrase,
				gOpt.SSHTimeout,
				gOpt.OptTimeout,
			).
			UserAction(host, globalOptions.User, globalOptions.Group, opt.SkipCreateUser || globalOptions.User == opt.User).
			EnvInit(host, globalOptions.User, globalOptions.Group).
			Mkdir(globalOptions.User, host, dirs...).
			BuildAsStep(fmt.Sprintf("  - Prepare %s:%d", host, info.ssh))
		envInitTasks = append(envInitTasks, t)
	}

	return envInitTasks
}

// buildDownloadCompTasks build download component tasks
func buildDownloadCompTasks(clusterVersion string, topo spec.Topology, logger *logprinter.Logger) []*task.StepDisplay {
	var tasks []*task.StepDisplay
	uniqueTasks := set.NewStringSet()

	topo.IterInstance(func(inst spec.Instance) {
		key := fmt.Sprintf("%s-%s-%s", inst.ComponentSource(), inst.OS(), inst.Arch())
		if found := uniqueTasks.Exist(key); !found {
			uniqueTasks.Insert(key)
			t := task.NewBuilder(logger).
				Download(inst.ComponentSource(), inst.OS(), inst.Arch(), clusterVersion).
				BuildAsStep(fmt.Sprintf("  - Download %s:%s (%s/%s)",
					inst.ComponentSource(), clusterVersion, inst.OS(), inst.Arch()))
			tasks = append(tasks, t)
		}
	})

	return tasks
}

// buildMkdirTasks builds the Mkdir tasks
func buildMkdirTasks(topo spec.Topology, gOpt *operator.Options, sshConnProps *gui.SSHConnectionProps, logger *logprinter.Logger) []*task.StepDisplay {
	base := topo.BaseTopo()
	globalOptions := base.GlobalOptions

	var mkdirCompTasks []*task.StepDisplay

	// Deploy components to remote
	topo.IterInstance(func(inst spec.Instance) {
		deployDir := spec.Abs(globalOptions.User, inst.DeployDir())
		// data dir would be empty for components which don't need it
		dataDirs := spec.Abs(globalOptions.User, inst.DataDir())
		// log dir will always be with values, but might not be used by the component
		logDir := spec.Abs(globalOptions.User, inst.LogDir())
		// Deploy component
		// prepare deployment server
		deployDirs := []string{
			deployDir, logDir,
			filepath.Join(deployDir, "bin"),
			filepath.Join(deployDir, "conf"),
			filepath.Join(deployDir, "scripts"),
		}

		t := task.NewBuilder(logger). // TODO: only support root deploy user
						RootSSH(
				inst.GetManageHost(),
				inst.GetSSHPort(),
				globalOptions.User,
				sshConnProps.Password,
				sshConnProps.IdentityFile,
				sshConnProps.IdentityFilePassphrase,
				gOpt.SSHTimeout,
				gOpt.OptTimeout,
			).
			//t := task.NewSimpleUerSSH(m.logger, inst.GetManageHost(), inst.GetSSHPort(), globalOptions.User, 0, 0).
			Mkdir(globalOptions.User, inst.GetManageHost(), deployDirs...).
			Mkdir(globalOptions.User, inst.GetManageHost(), dataDirs)

		mkdirCompTasks = append(mkdirCompTasks,
			t.BuildAsStep(fmt.Sprintf("  - Mkdir %s -> %s", strings.Join(append(deployDirs, dataDirs), ","), inst.GetHost())),
		)
	})

	return mkdirCompTasks
}

// buildDeployTasks builds the copy_component tasks
func buildDeployTasks(clusterName, clusterVersion string, topo spec.Topology, gOpt *operator.Options, sshConnProps *gui.SSHConnectionProps, logger *logprinter.Logger) []*task.StepDisplay {
	base := topo.BaseTopo()
	globalOptions := base.GlobalOptions

	var deployCompTasks []*task.StepDisplay

	uniqueHosts := getAllUniqueHosts(topo)
	var openGeminiComponentDeployTasksByHosts = make(map[string]*task.Builder, len(uniqueHosts)) // there are concurrent issues with the same node

	// Deploy components to remote
	topo.IterInstance(func(inst spec.Instance) {
		deployDir := spec.Abs(globalOptions.User, inst.DeployDir())

		tk, ok := openGeminiComponentDeployTasksByHosts[inst.GetHost()]
		if ok {
			tk = tk.CopyComponent(
				inst.ComponentSource(),
				inst.ComponentName(),
				inst.OS(),
				inst.Arch(),
				clusterVersion,
				"", // use default srcPath
				inst.GetManageHost(),
				deployDir,
			)
			openGeminiComponentDeployTasksByHosts[inst.GetHost()] = tk
			return
		}

		t := task.NewBuilder(logger). // TODO: only support root deploy user
						RootSSH(
				inst.GetManageHost(),
				inst.GetSSHPort(),
				globalOptions.User,
				sshConnProps.Password,
				sshConnProps.IdentityFile,
				sshConnProps.IdentityFilePassphrase,
				gOpt.SSHTimeout,
				gOpt.OptTimeout,
			)
		//t := task.NewSimpleUerSSH(m.logger, inst.GetManageHost(), inst.GetSSHPort(), globalOptions.User, 0, 0).

		if deployerInstance, ok := inst.(DeployerInstance); ok {
			deployerInstance.Deploy(t, "", deployDir, clusterVersion, clusterName, clusterVersion)
		} else {
			// copy dependency component if needed
			t = t.CopyComponent(
				inst.ComponentSource(),
				inst.ComponentName(),
				inst.OS(),
				inst.Arch(),
				clusterVersion,
				"", // use default srcPath
				inst.GetManageHost(),
				deployDir,
			)
		}
		// save task by host
		openGeminiComponentDeployTasksByHosts[inst.GetHost()] = t
	})

	for host, tk := range openGeminiComponentDeployTasksByHosts {
		deployCompTasks = append(deployCompTasks,
			tk.BuildAsStep(fmt.Sprintf("  - Copy %s -> %s", "required components", host)),
		)
	}

	return deployCompTasks
}

func buildInitConfigTasks(
	m *Manager,
	clustername string,
	topo spec.Topology,
	base *spec.BaseMeta,
	gOpt operator.Options,
) []*task.StepDisplay {
	var tasks []*task.StepDisplay

	topo.IterInstance(func(instance spec.Instance) {
		compName := instance.ComponentName()
		deployDir := spec.Abs(base.User, instance.DeployDir())
		// data dir would be empty for components which don't need it
		dataDir := spec.Abs(base.User, instance.DataDir())
		// log dir will always be with values, but might not be used by the component
		logDir := spec.Abs(base.User, instance.LogDir())

		// Download and copy the latest component to remote if the cluster is imported from Ansible
		tb := task.NewBuilder(m.logger)
		t := tb.InitConfig(
			clustername,
			base.Version,
			m.specManager,
			instance,
			base.User,
			gOpt.IgnoreConfigCheck,
			meta.DirPaths{
				Deploy: deployDir,
				Data:   dataDir,
				Log:    logDir,
				Cache:  m.specManager.Path(clustername, spec.TempConfigPath),
			},
		).
			BuildAsStep(fmt.Sprintf("  - Generate config %s -> %s", compName, instance.ID()))
		tasks = append(tasks, t)
	})

	return tasks
}
