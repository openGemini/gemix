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
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Toml struct {
	Common    CommonToml    `toml:"common"`
	Meta      MetaToml      `toml:"meta"`
	Http      HttpToml      `toml:"http"`
	Data      DataToml      `toml:"data"`
	Logging   LoggingToml   `toml:"logging"`
	Gossip    GossipToml    `toml:"gossip"`
	Monitor   MonitorToml   `toml:"monitor"`
	Retention RetentionToml `toml:"retention"`
}

type CommonToml struct {
	MetaJoin []string `toml:"meta-join"`
	HaPolicy string   `toml:"ha-policy"`
}

type MetaToml struct {
	BindAddress     string `toml:"bind-address"`
	HttpBindAddress string `toml:"http-bind-address"`
	RpcBindAddress  string `toml:"rpc-bind-address"`
	Dir             string `toml:"dir"`
}

type HttpToml struct {
	BindAddress string `toml:"bind-address"`

	AuthEnabled      bool   `toml:"auth-enabled"`
	HttpsEnabled     bool   `toml:"https-enabled"`
	HttpsCertificate string `toml:"https-certificate"`
	HttpsPrivateKey  string `toml:"https-private-key"`
}

type DataToml struct {
	StoreIngestAddr     string `toml:"store-ingest-addr"`
	StoreSelectAddr     string `toml:"store-select-addr"`
	StoreDataDir        string `toml:"store-data-dir"`
	StoreWalDir         string `toml:"store-wal-dir"`
	StoreMetaDir        string `toml:"store-meta-dir"`
	CacheTableDataBlock bool   `toml:"cache-table-data-block"`
	CacheTableMetaBlock bool   `toml:"cache-table-meta-block"`
	ReadCacheLimit      int    `toml:"read-cache-limit"`
}

type LoggingToml struct {
	Path string `toml:"path"`
}

type GossipToml struct {
	BindAddress   string   `toml:"bind-address"`
	StoreBindPort int      `toml:"store-bind-port"`
	MetaBindPort  int      `toml:"meta-bind-port"`
	Members       []string `toml:"members"`
}

type MonitorToml struct {
	Pushers       string `toml:"pushers"`
	StoreEnabled  bool   `toml:"store-enabled"`
	StoreDatabase string `toml:"store-interval"`
	StoreInterval string `toml:"store-database"`
	StorePath     string `toml:"store-path"`
	Compress      bool   `toml:"compress"`
	HttpEndpoint  string `toml:"http-endpoint"`
	Username      string `toml:"username"`
	Password      string `toml:"password"`
}

type RetentionToml struct {
	CheckInterval string `toml:"check-interval"`
}

func ReadFromToml(tomlPath string) (Toml, error) {
	var t Toml
	if _, err := toml.DecodeFile(tomlPath, &t); err != nil {
		return t, err
	}
	return t, nil
}

func ConvertToml(hostToml Toml, pwd string) Toml {
	if len(hostToml.Meta.Dir) > 1 && hostToml.Meta.Dir[:1] == "~" {
		hostToml.Meta.Dir = filepath.Join(pwd, hostToml.Meta.Dir[1:])
	}
	if len(hostToml.Data.StoreDataDir) > 1 && hostToml.Data.StoreDataDir[:1] == "~" {
		hostToml.Data.StoreDataDir = filepath.Join(pwd, hostToml.Data.StoreDataDir[1:])
	}
	if len(hostToml.Data.StoreWalDir) > 1 && hostToml.Data.StoreWalDir[:1] == "~" {
		hostToml.Data.StoreWalDir = filepath.Join(pwd, hostToml.Data.StoreWalDir[1:])
	}
	if len(hostToml.Data.StoreMetaDir) > 1 && hostToml.Data.StoreMetaDir[:1] == "~" {
		hostToml.Data.StoreMetaDir = filepath.Join(pwd, hostToml.Data.StoreMetaDir[1:])
	}
	if len(hostToml.Logging.Path) > 1 && hostToml.Logging.Path[:1] == "~" {
		hostToml.Logging.Path = filepath.Join(pwd, hostToml.Logging.Path[1:])
	}
	return hostToml
}

func GenNewToml(t Toml, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0640)
	if err != nil {
		return err
	}
	defer f.Close()

	e := toml.NewEncoder(f)
	return e.Encode(t)
}
