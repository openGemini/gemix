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

package spec

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestGlobalOptions(t *testing.T) {
	topo := Specification{}
	err := yaml.Unmarshal([]byte(`
global:
  user: "test1"
  ssh_port: 220
  deploy_dir: "test-deploy"
  data_dir: "test-data"
ts_meta_servers:
  - host: 172.16.5.138
    deploy_dir: "gemini-deploy"
ts_store_servers:
  - host: 172.16.5.53
    data_dir: "ts-store-data"
ts_sql_servers:
  - host: 172.16.5.233
    data_dir: "ts-sql-data"
`), &topo)
	assert.NoError(t, err)
	assert.Equal(t, "test1", topo.GlobalOptions.User)
	assert.Equal(t, 220, topo.GlobalOptions.SSHPort)

	assert.Equal(t, 220, topo.TSMetaServers[0].SSHPort)
	assert.Equal(t, "gemini-deploy", topo.TSMetaServers[0].DeployDir)
	assert.Equal(t, "test-data", topo.TSMetaServers[0].DataDir)

	assert.Equal(t, 220, topo.TSStoreServers[0].SSHPort)
	assert.Equal(t, filepath.Join("test-deploy", "ts-store-8401"), topo.TSStoreServers[0].DeployDir)
	assert.Equal(t, "ts-store-data", topo.TSStoreServers[0].DataDir)

	assert.Equal(t, 220, topo.TSSqlServers[0].SSHPort)
	assert.Equal(t, filepath.Join("test-deploy", "ts-sql-8086"), topo.TSSqlServers[0].DeployDir)
	assert.Equal(t, filepath.Join("logs"), topo.TSSqlServers[0].LogDir)
}
