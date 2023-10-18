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

type CommonConfig struct {
	MetaHosts  []string //IPs
	StoreHosts []string
	SqlHosts   []string
	User       string
	Os         string
	Arch       string
}

type SSHConfig struct {
	// get from yaml
	Port       int
	UpDataPath string
	LogPath    string
}

type Config struct {
	CommonConfig *CommonConfig
	SSHConfig    map[string]SSHConfig
}

type Configurator interface {
	GenClusterConfs() error
	BuildConfig() error
	GetConfig() *Config
}

type GeminiConfigurator struct {
	yamlPath string
	tomlPath string
	genPath  string
	conf     *Config
	yaml     *Yaml
}

func NewGeminiConfigurator(yPath, tPath, gPath string) Configurator {
	return &GeminiConfigurator{
		yamlPath: yPath,
		tomlPath: tPath,
		genPath:  gPath,
		conf: &Config{
			CommonConfig: &CommonConfig{},
			SSHConfig:    make(map[string]SSHConfig),
		},
	}
}

func (c *GeminiConfigurator) GenClusterConfs() error {
	var err error
	var t Toml
	// generate new toml files
	if t, err = ReadFromToml(c.tomlPath); err != nil {
		return err
	}
	return GenConfs(*c.yaml, t, c.genPath)
}

func (c *GeminiConfigurator) BuildConfig() error {
	var err error
	var y Yaml
	if y, err = ReadFromYaml(c.yamlPath); err != nil {
		return err
	}
	c.buildFromYaml(y)
	c.yaml = &y
	return err
}

func (c *GeminiConfigurator) GetConfig() *Config {
	return c.conf
}

func (c *GeminiConfigurator) buildFromYaml(y Yaml) {
	c.conf.CommonConfig.User = y.Global.User
	c.conf.CommonConfig.Os = y.Global.OS
	c.conf.CommonConfig.Arch = y.Global.Arch

	for _, meta := range y.TsMeta {
		ssh, ok := c.conf.SSHConfig[meta.Host]
		if !ok {
			ssh = SSHConfig{}
		}
		if meta.SSHPort != 0 {
			ssh.Port = meta.SSHPort
		}
		if meta.DeployDir != "" {
			ssh.UpDataPath = meta.DeployDir
		}
		if meta.LogDir != "" {
			ssh.LogPath = meta.LogDir
		}
		c.conf.SSHConfig[meta.Host] = ssh

		c.conf.CommonConfig.MetaHosts = append(c.conf.CommonConfig.MetaHosts, meta.Host)
	}
	for _, sql := range y.TsSql {
		ssh, ok := c.conf.SSHConfig[sql.Host]
		if !ok {
			ssh = SSHConfig{}
		}
		if sql.SSHPort != 0 {
			ssh.Port = sql.SSHPort
		}
		if sql.DeployDir != "" {
			ssh.UpDataPath = sql.DeployDir
		}
		if sql.LogDir != "" {
			ssh.LogPath = sql.LogDir
		}
		c.conf.SSHConfig[sql.Host] = ssh

		c.conf.CommonConfig.SqlHosts = append(c.conf.CommonConfig.SqlHosts, sql.Host)
	}
	for _, store := range y.TsStore {
		ssh, ok := c.conf.SSHConfig[store.Host]
		if !ok {
			ssh = SSHConfig{}
		}
		if store.SSHPort != 0 {
			ssh.Port = store.SSHPort
		}
		if store.DeployDir != "" {
			ssh.UpDataPath = store.DeployDir
		}
		if store.LogDir != "" {
			ssh.LogPath = store.LogDir
		}
		c.conf.SSHConfig[store.Host] = ssh

		c.conf.CommonConfig.StoreHosts = append(c.conf.CommonConfig.StoreHosts, store.Host)
	}
}
