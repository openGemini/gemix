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

// checkConflict checks cluster conflict
func checkConflict(m *Manager, clusterName string, topo *spec.Specification) error {
	//clusterList, err := m.specManager.GetAllClusters()
	//if err != nil {
	//	return err
	//}
	//// use a dummy cluster name, the real cluster name is set during deploy
	//if err := spec.CheckClusterPortConflict(clusterList, clusterName, topo); err != nil {
	//	return err
	//}
	//err = spec.CheckClusterDirConflict(clusterList, clusterName, topo)
	return nil
}