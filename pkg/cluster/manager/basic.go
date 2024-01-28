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
	"path/filepath"

	"github.com/openGemini/gemix/pkg/cluster/spec"
	"github.com/openGemini/gemix/pkg/set"
	"github.com/openGemini/gemix/pkg/utils"
)

// getAllUniqueHosts gets all the instance
func getAllUniqueHosts(topo spec.Topology) map[string]spec.MonitorHostInfo {
	// monitor
	uniqueHosts := make(map[string]spec.MonitorHostInfo) // host -> ssh-port, os, arch
	topo.IterInstance(func(inst spec.Instance) {
		if _, found := uniqueHosts[inst.GetManageHost()]; !found {
			uniqueHosts[inst.GetManageHost()] = spec.MonitorHostInfo{
				Ssh:  inst.GetSSHPort(),
				Os:   inst.OS(),
				Arch: inst.Arch(),
			}
		}
	})

	return uniqueHosts
}

// getMonitorHosts  get the ts-monitor instances
func getMonitorHosts(topo spec.Topology) (map[string]*spec.MonitorHostInfo, set.StringSet) {
	// monitor
	uniqueHosts := make(map[string]*spec.MonitorHostInfo) // host -> ssh-port, os, arch
	noAgentHosts := set.NewStringSet()
	topo.IterInstance(func(inst spec.Instance) {
		// add the instance to ignore list if it marks itself as ignore_exporter
		//if inst.IgnoreMonitorAgent() {
		//	noAgentHosts.Insert(inst.GetManageHost())
		//}

		if h, found := uniqueHosts[inst.GetManageHost()]; !found {
			uniqueHosts[inst.GetManageHost()] = &spec.MonitorHostInfo{
				Ssh:          inst.GetSSHPort(),
				Os:           inst.OS(),
				Arch:         inst.Arch(),
				MetricPath:   filepath.Join(inst.LogDir(), "metric"),
				ErrorLogPath: inst.LogDir(),
				ProcessName:  set.NewStringSet(),
			}
		} else {
			switch inst.ComponentName() {
			case spec.ComponentTSMeta,
				spec.ComponentTSStore,
				spec.ComponentTSSql:
				h.ProcessName.Insert(inst.ComponentName())
			default:
			}
		}
		h := uniqueHosts[inst.GetManageHost()]
		if inst.ComponentName() == spec.ComponentTSStore {
			h.DataPath = inst.DataDir()
			h.WALPath = filepath.Join(inst.DataDir(), "wal")
		}

		if inst.ComponentName() == spec.ComponentTSServer {
			h.MonitorAddr = utils.JoinHostPort(inst.GetHost(), inst.GetPort())
		}

	})

	return uniqueHosts, noAgentHosts
}
