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
	Global       GlobalYaml   `yaml:"global"`
	ServerConfig ServerConfig `yaml:"server_config"`
	TsMeta       []MetaYaml   `yaml:"ts-meta"`
	TsSql        []SqlYaml    `yaml:"ts-sql"`
	TsStore      []StoreYaml  `yaml:"ts-store"`
}

type GlobalYaml struct {
	SSHPort   int    `yaml:"ssh_port"`
	User      string `yaml:"user"`
	LogDir    string `yaml:"log_dir"`
	DeployDir string `yaml:"deploy_dir"`

	OS   string `yaml:"os"`
	Arch string `yaml:"arch"`
}

type ServerConfig struct {
	TsMeta  MetaYamlConfig  `yaml:"ts-meta"`
	TsSql   SqlYamlConfig   `yaml:"ts-sql"`
	TsStore StoreYamlConfig `yaml:"ts-store"`
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

	Config MetaYamlConfig `yaml:"config"`
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

	Config SqlYamlConfig `yaml:"config"`
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

	Config StoreYamlConfig `yaml:"config"`
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
type SqlYamlConfig struct {
	AuthEnabled      bool   `yaml:"http.auth-enabled"`
	HttpsEnabled     bool   `yaml:"http.https-enabled"`
	HttpsCertificate string `yaml:"http.https-certificate"`
	HttpsPrivateKey  string `yaml:"http.https-private-key"`

	HaPolicy string `yaml:"common.ha-policy"`

	CheckInterval string `yaml:"retention.check-interval"`

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
type StoreYamlConfig struct {
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

		if y.TsMeta[i].Config.HaPolicy == "" {
			y.TsMeta[i].Config.HaPolicy = y.ServerConfig.TsMeta.HaPolicy
		}
		if y.TsMeta[i].Config.Pushers == "" {
			y.TsMeta[i].Config.Pushers = y.ServerConfig.TsMeta.Pushers
		}
		if y.TsMeta[i].Config.StoreDatabase == "" {
			y.TsMeta[i].Config.StoreDatabase = y.ServerConfig.TsMeta.StoreDatabase
		}
		if y.TsMeta[i].Config.StoreInterval == "" {
			y.TsMeta[i].Config.StoreInterval = y.ServerConfig.TsMeta.StoreInterval
		}
		if y.TsMeta[i].Config.StorePath == "" {
			y.TsMeta[i].Config.StorePath = y.ServerConfig.TsMeta.StorePath
		}
		if y.TsMeta[i].Config.HttpEndpoint == "" {
			y.TsMeta[i].Config.HttpEndpoint = y.ServerConfig.TsMeta.HttpEndpoint
		}
		if y.TsMeta[i].Config.Username == "" {
			y.TsMeta[i].Config.Username = y.ServerConfig.TsMeta.Username
		}
		if y.TsMeta[i].Config.Password == "" {
			y.TsMeta[i].Config.Password = y.ServerConfig.TsMeta.Password
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

		if y.TsSql[i].Config.HttpsCertificate == "" {
			y.TsSql[i].Config.HttpsCertificate = y.ServerConfig.TsSql.HttpsCertificate
		}
		if y.TsSql[i].Config.HttpsPrivateKey == "" {
			y.TsSql[i].Config.HttpsPrivateKey = y.ServerConfig.TsSql.HttpsPrivateKey
		}
		if y.TsSql[i].Config.CheckInterval == "" {
			y.TsSql[i].Config.CheckInterval = y.ServerConfig.TsSql.CheckInterval
		}

		if y.TsSql[i].Config.HaPolicy == "" {
			y.TsSql[i].Config.HaPolicy = y.ServerConfig.TsSql.HaPolicy
		}
		if y.TsSql[i].Config.Pushers == "" {
			y.TsSql[i].Config.Pushers = y.ServerConfig.TsSql.Pushers
		}
		if y.TsSql[i].Config.StoreDatabase == "" {
			y.TsSql[i].Config.StoreDatabase = y.ServerConfig.TsMeta.StoreDatabase
		}
		if y.TsSql[i].Config.StoreInterval == "" {
			y.TsSql[i].Config.StoreInterval = y.ServerConfig.TsSql.StoreInterval
		}
		if y.TsSql[i].Config.StorePath == "" {
			y.TsSql[i].Config.StorePath = y.ServerConfig.TsSql.StorePath
		}
		if y.TsSql[i].Config.HttpEndpoint == "" {
			y.TsSql[i].Config.HttpEndpoint = y.ServerConfig.TsSql.HttpEndpoint
		}
		if y.TsSql[i].Config.Username == "" {
			y.TsSql[i].Config.Username = y.ServerConfig.TsSql.Username
		}
		if y.TsSql[i].Config.Password == "" {
			y.TsSql[i].Config.Password = y.ServerConfig.TsSql.Password
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

		if y.TsStore[i].Config.HaPolicy == "" {
			y.TsStore[i].Config.HaPolicy = y.ServerConfig.TsStore.HaPolicy
		}
		if y.TsStore[i].Config.Pushers == "" {
			y.TsStore[i].Config.Pushers = y.ServerConfig.TsStore.Pushers
		}
		if y.TsStore[i].Config.StoreDatabase == "" {
			y.TsStore[i].Config.StoreDatabase = y.ServerConfig.TsStore.StoreDatabase
		}
		if y.TsStore[i].Config.StoreInterval == "" {
			y.TsStore[i].Config.StoreInterval = y.ServerConfig.TsStore.StoreInterval
		}
		if y.TsStore[i].Config.StorePath == "" {
			y.TsStore[i].Config.StorePath = y.ServerConfig.TsStore.StorePath
		}
		if y.TsStore[i].Config.HttpEndpoint == "" {
			y.TsStore[i].Config.HttpEndpoint = y.ServerConfig.TsStore.HttpEndpoint
		}
		if y.TsStore[i].Config.Username == "" {
			y.TsStore[i].Config.Username = y.ServerConfig.TsStore.Username
		}
		if y.TsStore[i].Config.Password == "" {
			y.TsStore[i].Config.Password = y.ServerConfig.TsStore.Password
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
