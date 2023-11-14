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
	"github.com/creasty/defaults"
	"github.com/pkg/errors"
)

type (
	// GlobalOptions represents the global options for all groups in topology
	// specification in topology.yaml
	GlobalOptions struct {
		User    string `yaml:"user,omitempty" default:"gemini"`
		Group   string `yaml:"group,omitempty"`
		SSHPort int    `yaml:"ssh_port,omitempty" default:"22" validate:"ssh_port:editable"`
		//SSHType    executor.SSHType `yaml:"ssh_type,omitempty" default:"builtin"`
		TLSEnabled bool   `yaml:"enable_tls,omitempty"`
		DeployDir  string `yaml:"deploy_dir,omitempty" default:"deploy"`
		//DataDir         string               `yaml:"data_dir,omitempty" default:"data"`
		LogDir string `yaml:"log_dir,omitempty"`
		//ResourceControl meta.ResourceControl `yaml:"resource_control,omitempty" validate:"resource_control:editable"`
		OS   string `yaml:"os,omitempty" default:"linux"`
		Arch string `yaml:"arch,omitempty"`
		//Custom          any                  `yaml:"custom,omitempty" validate:"custom:ignore"`
	}

	// ServerConfigs represents the server runtime configuration
	ServerConfigs struct {
		TsMeta  map[string]any `yaml:"ts-meta"`
		TsSql   map[string]any `yaml:"ts-sql"`
		TsStore map[string]any `yaml:"ts-store"`
	}

	// Specification represents the specification of topology.yaml
	Specification struct {
		GlobalOptions GlobalOptions `yaml:"global,omitempty" validate:"global:editable"`
		//MonitoredOptions MonitoredOptions `yaml:"monitored,omitempty" validate:"monitored:editable"`
		ServerConfigs ServerConfigs `yaml:"server_configs,omitempty" validate:"server_configs:ignore"`
		TSMetaServers []*TSMetaSpec `yaml:"ts_meta_servers"`
		//TSSqlServers   []*TSSqlSpec   `yaml:"ts_sql_servers"`
		//TSStoreServers []*TSStoreSpec `yaml:"ts_store_servers"`
		//DashboardServers []*DashboardSpec `yaml:"opengemini_dashboard_servers,omitempty"`
		//Monitors         []*PrometheusSpec    `yaml:"monitoring_servers"`
		//Grafanas         []*GrafanaSpec       `yaml:"grafana_servers,omitempty"`

	}
)

// UnmarshalYAML implements the yaml.Unmarshaler interface,
// it sets the default values when unmarshaling the topology file
func (s *Specification) UnmarshalYAML(unmarshal func(any) error) error {
	type topology Specification
	if err := unmarshal((*topology)(s)); err != nil {
		return err
	}

	// set default values from tag
	if err := defaults.Set(s); err != nil {
		return errors.WithStack(err)
	}

	// populate custom default values as needed
	//if err := fillCustomDefaults(&s.GlobalOptions, s); err != nil {
	//	return err
	//}

	//return s.Validate()
	return nil
}
