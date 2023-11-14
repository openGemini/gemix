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

// TSMetaSpec represents the ts-meta topology specification in topology.yaml
type TSMetaSpec struct {
	Host string `yaml:"host"`
	// default value in GlobalYaml
	SSHPort   int    `yaml:"ssh_port,omitempty" validate:"ssh_port:editable"`
	LogDir    string `yaml:"log_dir"`
	DeployDir string `yaml:"deploy_dir"`
	DataDir   string `yaml:"data_dir,omitempty"`

	// port specification
	ClientPort int `yaml:"client_port" default:"8091"`
	PeerPort   int `yaml:"peer_port" default:"8092"`
	RaftPort   int `yaml:"raft_port"`
	GossipPort int `yaml:"gossip_port"`

	Config MetaYamlConfig `yaml:"config,omitempty" validate:"config:ignore"`
}

type MetaYamlConfig struct {
	HaPolicy string `yaml:"common.ha-policy"`

	Pushers       string `yaml:"monitor.pushers"`
	StoreEnabled  bool   `yaml:"monitor.store-enabled"`
	StoreDatabase string `yaml:"monitor.store-interval"`
	StoreInterval string `yaml:"monitor.store-database"`
	StorePath     string `yaml:"monitor.store-path"`
	Compress      bool   `yaml:"monitor.compress"`
	HttpEndpoint  string `yaml:"monitor.http-endpoint"`
	Username      string `yaml:"monitor.username"`
	Password      string `yaml:"monitor.password"`
}
