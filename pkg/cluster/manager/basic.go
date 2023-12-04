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

import "github.com/openGemini/gemix/pkg/cluster/spec"

type hostInfo struct {
	ssh  int    // ssh port of host
	os   string // operating system
	arch string // cpu architecture
}

// getAllUniqueHosts gets all the instance
func getAllUniqueHosts(topo spec.Topology) map[string]hostInfo {
	// monitor
	uniqueHosts := make(map[string]hostInfo) // host -> ssh-port, os, arch
	topo.IterInstance(func(inst spec.Instance) {
		if _, found := uniqueHosts[inst.GetManageHost()]; !found {
			uniqueHosts[inst.GetManageHost()] = hostInfo{
				ssh:  inst.GetSSHPort(),
				os:   inst.OS(),
				arch: inst.Arch(),
			}
		}
	})

	return uniqueHosts
}
