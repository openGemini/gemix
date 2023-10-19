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

package config

import (
	"errors"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Yaml struct {
	Global  GlobalYaml  `yaml:"global"`
	TsMeta  []MetaYaml  `yaml:"ts-meta"`
	TsSql   []SqlYaml   `yaml:"ts-sql"`
	TsStore []StoreYaml `yaml:"ts-store"`
}

type GlobalYaml struct {
	SSHPort   int    `yaml:"ssh_port"`
	User      string `yaml:"user"`
	LogDir    string `yaml:"log_dir"`
	DeployDir string `yaml:"deploy_dir"`

	OS   string `yaml:"os"`
	Arch string `yaml:"arch"`
}

type MetaYaml struct {
	Host string `yaml:"host"`
	// default value in GlobalYaml
	SSHPort   int    `yaml:"ssh_port"`
	LogDir    string `yaml:"log_dir"`
	DeployDir string `yaml:"deploy_dir"`

	// do not need default value
	ClientPort int    `yaml:"client_port"`
	PeerPort   int    `yaml:"peer_port"`
	RaftPort   int    `yaml:"raft_port"`
	GossipPort int    `yaml:"gossip_port"`
	DataDir    string `yaml:"data_dir"`
}

type SqlYaml struct {
	Host string `yaml:"host"`
	// default value in GlobalYaml
	SSHPort   int    `yaml:"ssh_port"`
	LogDir    string `yaml:"log_dir"`
	DeployDir string `yaml:"deploy_dir"`

	// do not need default value
	Port       int `yaml:"port"`
	FlightPort int `yaml:"flight_port"`
}

type StoreYaml struct {
	Host string `yaml:"host"`
	// default value in GlobalYaml
	SSHPort   int    `yaml:"ssh_port"`
	LogDir    string `yaml:"log_dir"`
	DeployDir string `yaml:"deploy_dir"`

	// do not need default value
	IngestPort int    `yaml:"ingest_port"`
	SelectPort int    `yaml:"select_port"`
	GossipPort int    `yaml:"gossip_port"`
	DataDir    string `yaml:"data_dir"`
	MetaDir    string `yaml:"meta_dir"`
}

func checkRequiredOptions(y Yaml) bool {
	if y.Global.OS == "" || y.Global.Arch == "" {
		return false
	}
	if y.Global.LogDir == "" || y.Global.DeployDir == "" || y.Global.SSHPort == 0 {
		return false
	}

	for _, meta := range y.TsMeta {
		if meta.Host == "" || meta.DataDir == "" {
			return false
		}
	}
	for _, sql := range y.TsSql {
		if sql.Host == "" {
			return false
		}
	}
	for _, store := range y.TsStore {
		if store.Host == "" || store.DataDir == "" || store.MetaDir == "" {
			return false
		}
	}

	return true
}

func updataWithGlobalDefaults(y *Yaml) {
	for i := range y.TsMeta {
		if y.TsMeta[i].SSHPort == 0 {
			y.TsMeta[i].SSHPort = y.Global.SSHPort
		}
		if y.TsMeta[i].LogDir == "" {
			y.TsMeta[i].LogDir = y.Global.LogDir
		}
		if y.TsMeta[i].DeployDir == "" {
			y.TsMeta[i].DeployDir = y.Global.DeployDir
		}
	}
	for i := range y.TsSql {
		if y.TsSql[i].SSHPort == 0 {
			y.TsSql[i].SSHPort = y.Global.SSHPort
		}
		if y.TsSql[i].LogDir == "" {
			y.TsSql[i].LogDir = y.Global.LogDir
		}
		if y.TsSql[i].DeployDir == "" {
			y.TsSql[i].DeployDir = y.Global.DeployDir
		}
	}
	for i := range y.TsStore {
		if y.TsStore[i].SSHPort == 0 {
			y.TsStore[i].SSHPort = y.Global.SSHPort
		}
		if y.TsStore[i].LogDir == "" {
			y.TsStore[i].LogDir = y.Global.LogDir
		}
		if y.TsStore[i].DeployDir == "" {
			y.TsStore[i].DeployDir = y.Global.DeployDir
		}
	}
}

func ReadFromYaml(yamlPath string) (Yaml, error) {
	var err error
	yamlFile, err := ioutil.ReadFile(yamlPath)
	if err != nil {
		return Yaml{}, err
	}
	var y Yaml
	if err = yaml.Unmarshal(yamlFile, &y); err != nil {
		return Yaml{}, err
	}

	// check required options
	if pass := checkRequiredOptions(y); !pass {
		return Yaml{}, errors.New("missing requitred options for yaml configuration file")
	}

	// Update with default values
	updataWithGlobalDefaults(&y)

	return y, nil
}
