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
	"context"
	"fmt"
	"path/filepath"

	"github.com/openGemini/gemix/pkg/cluster/ctxt"
	"github.com/openGemini/gemix/pkg/cluster/template/scripts"
	"github.com/openGemini/gemix/pkg/meta"
	"github.com/openGemini/gemix/pkg/utils"
	"github.com/pkg/errors"
)

// TSSqlSpec represents the ts-sql topology specification in topology.yaml
type TSSqlSpec struct {
	Arch string `yaml:"arch,omitempty"`
	OS   string `yaml:"os,omitempty"`

	Source string `yaml:"source,omitempty" validate:"source:editable"`

	// Use Name to get the name with a default value if it's empty.
	Name string `yaml:"name"`

	Host       string `yaml:"host"`
	ManageHost string `yaml:"manage_host,omitempty" validate:"manage_host:editable"`
	ListenHost string `yaml:"listen_host,omitempty"`
	SSHPort    int    `yaml:"ssh_port,omitempty" validate:"ssh_port:editable"`

	LogDir    string `yaml:"log_dir"`
	DeployDir string `yaml:"deploy_dir"`

	// port specification
	Port int `yaml:"port" default:"8086"`
	// FlightPort int `yaml:"flight_port"  default:"8087"` // define at ts-data

	Config map[string]any `yaml:"config,omitempty" validate:"config:ignore"`
}

func (s *TSSqlSpec) SSH() (string, int) {
	host := s.Host
	if s.ManageHost != "" {
		host = s.ManageHost
	}
	return host, s.SSHPort
}

// Role returns the component role of the instance
func (s *TSSqlSpec) Role() string {
	return ComponentTSSql
}

// GetManageHost returns the manage host of the instance
func (s *TSSqlSpec) GetManageHost() string {
	if s.ManageHost != "" {
		return s.ManageHost
	}
	return s.Host
}

// GetSource returns source to download the component
func (s *TSSqlSpec) GetSource() string {
	if s.Source == "" {
		return ComponentOpenGemini
	}
	return s.Source
}

// TSSqlComponent represents ts-meta component.
type TSSqlComponent struct{ Topology *Specification }

// Name implements Component interface.
func (c *TSSqlComponent) Name() string {
	return ComponentTSSql
}

// Role implements Component interface.
func (c *TSSqlComponent) Role() string {
	return ComponentTSSql
}

// Instances implements Component interface.
func (c *TSSqlComponent) Instances() []Instance {
	ins := make([]Instance, 0, len(c.Topology.TSSqlServers))
	for _, s := range c.Topology.TSSqlServers {
		s := s
		ins = append(ins, &TSSqlInstance{
			Name: c.Name(),
			BaseInstance: BaseInstance{
				InstanceSpec: s,
				Name:         c.Name(),
				Host:         s.Host,
				ManageHost:   s.ManageHost,
				ListenHost:   s.ListenHost,
				Port:         s.Port,
				SSHP:         s.SSHPort,
				Source:       s.GetSource(),

				Ports: []int{
					s.Port,
					//s.FlightPort,
				},
				Dirs: []string{
					s.DeployDir,
					s.LogDir,
				},
				//StatusFn: s.Status,
				//UptimeFn: func(_ context.Context, timeout time.Duration, tlsCfg *tls.Config) time.Duration {
				//	return UptimeByHost(s.GetManageHost(), s.ClientPort, timeout, tlsCfg)
				//},
			},
			topo: c.Topology,
		})
	}
	return ins
}

// TSSqlInstance represent the ts-meta instance
type TSSqlInstance struct {
	Name string
	BaseInstance
	topo *Specification
}

func (i *TSSqlInstance) InitConfig(ctx context.Context, e ctxt.Executor, clusterName string, clusterVersion string, deployUser string, paths meta.DirPaths) error {
	topo := i.topo
	if err := i.BaseInstance.InitConfig(ctx, e, topo.GlobalOptions, deployUser, paths); err != nil {
		return err
	}
	spec := i.InstanceSpec.(*TSSqlSpec)

	cfg := &scripts.TSSqlScript{
		DeployDir: paths.Deploy,
		LogDir:    paths.Log,
	}

	fp := filepath.Join(paths.Cache, fmt.Sprintf("run_ts_sql_%s_%d.sh", i.GetHost(), i.GetPort()))
	if err := cfg.ConfigToFile(fp); err != nil {
		return errors.WithStack(err)
	}

	dst := filepath.Join(paths.Deploy, "scripts", "run_ts-sql.sh")
	if err := e.Transfer(ctx, fp, dst, false, 0, false); err != nil {
		return errors.WithStack(err)
	}
	_, _, err := e.Execute(ctx, "chmod +x "+dst, false)
	if err != nil {
		return errors.WithStack(err)
	}

	globalConfig := topo.ServerConfigs.TsSql

	// set TLS configs
	//spec.Config, err = i.setTLSConfig(ctx, enableTLS, spec.Config, paths)
	//if err != nil {
	//	return errors.WithStack(err)
	//}

	configs := i.SetDefaultConfig(spec.Config)

	if err = i.MergeServerConfig(ctx, e, globalConfig, configs, paths); err != nil {
		return errors.WithStack(err)
	}

	return checkConfig(ctx, e, i.ComponentName(), i.ComponentSource(), clusterVersion, i.OS(), i.Arch(), i.ComponentName()+".toml", paths, nil)
}

func (i *TSSqlInstance) SetDefaultConfig(instanceConf map[string]any) map[string]any {
	if instanceConf == nil {
		instanceConf = make(map[string]any, 20)
	}
	var metaPeerAddrs []string
	for _, metaSpec := range i.topo.TSMetaServers {
		metaPeerAddrs = append(metaPeerAddrs, utils.JoinHostPort(metaSpec.Host, metaSpec.PeerPort))
	}
	instanceConf["common.meta-join"] = metaPeerAddrs

	var tsSqlSpec *TSSqlSpec
	for _, sqlSpec := range i.topo.TSSqlServers {
		if i.Host == sqlSpec.Host {
			tsSqlSpec = sqlSpec
		}
	}
	instanceConf["http.bind-address"] = utils.JoinHostPort(i.Host, tsSqlSpec.Port)
	instanceConf["http.flight-enabled"] = false // enabled at ts-data
	instanceConf["logging.path"] = tsSqlSpec.LogDir

	return instanceConf
}
