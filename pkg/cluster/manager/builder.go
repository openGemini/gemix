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

	operator "github.com/openGemini/gemix/pkg/cluster/operation"
	"github.com/openGemini/gemix/pkg/cluster/spec"
	"github.com/openGemini/gemix/pkg/cluster/task"
	"github.com/openGemini/gemix/pkg/meta"
	"go.uber.org/zap"
)

// buildDownloadCompTasks build download component tasks
func buildDownloadCompTasks(clusterVersion string, topo *spec.Specification) []*task.StepDisplay {
	var tasks []*task.StepDisplay
	uniqueTasks := make(map[string]struct{})

	topo.IterInstance(func(inst spec.Instance) {
		key := fmt.Sprintf("%s-%s-%s", inst.ComponentSource(), inst.OS(), inst.Arch())
		if _, found := uniqueTasks[key]; !found {
			uniqueTasks[key] = struct{}{}

			t := task.NewBuilder(zap.L()).
				Download(inst.ComponentSource(), inst.OS(), inst.Arch(), clusterVersion).
				BuildAsStep(fmt.Sprintf("  - Download %s:%s (%s/%s)",
					inst.ComponentSource(), clusterVersion, inst.OS(), inst.Arch()))
			tasks = append(tasks, t)
		}
	})

	return tasks
}

func buildInitConfigTasks(
	m *Manager,
	clustername string,
	topo *spec.Specification,
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
