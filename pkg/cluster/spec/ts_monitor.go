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
	"crypto/tls"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/openGemini/gemix/pkg/cluster/ctxt"
	"github.com/openGemini/gemix/pkg/cluster/template/scripts"
	"github.com/openGemini/gemix/pkg/meta"
	"github.com/openGemini/gemix/pkg/set"
	"github.com/pkg/errors"
)

// TSMonitorSpec represents the ts_monitor_enabled topology specification in topology.yaml
type TSMonitorSpec struct {
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

	MonitorProcess map[string]struct{}
	Config         map[string]any `yaml:"config,omitempty" validate:"config:ignore"`
}

func (s *TSMonitorSpec) SSH() (string, int) {
	host := s.Host
	if s.ManageHost != "" {
		host = s.ManageHost
	}
	return host, s.SSHPort
}

// Role returns the component role of the instance
func (s *TSMonitorSpec) Role() string {
	return ComponentTSMonitor
}

// GetManageHost returns the manage host of the instance
func (s *TSMonitorSpec) GetManageHost() string {
	if s.ManageHost != "" {
		return s.ManageHost
	}
	return s.Host
}

// GetSource returns source to download the component
func (s *TSMonitorSpec) GetSource() string {
	if s.Source == "" {
		return ComponentOpenGemini
	}
	return s.Source
}

// TSMonitorComponent represents ts-meta component.
type TSMonitorComponent struct{ Topology *Specification }

// Name implements Component interface.
func (c *TSMonitorComponent) Name() string {
	return ComponentTSMonitor
}

// Role implements Component interface.
func (c *TSMonitorComponent) Role() string {
	return ComponentTSMonitor
}

// Instances implements Component interface.
func (c *TSMonitorComponent) Instances() []Instance {
	if !c.Topology.MonitoredOptions.TSMonitorEnabled {
		return nil
	}

	instanceSet := set.NewStringSet()

	ins := make([]Instance, 0, len(c.Topology.TSMetaServers)+len(c.Topology.TSSqlServers)+len(c.Topology.TSStoreServers))
	for _, s := range c.Topology.TSStoreServers {
		if instanceSet.Exist(s.Host) {
			continue
		}
		ms := &TSMonitorSpec{
			OS:        s.OS,
			Arch:      s.Arch,
			DeployDir: c.Topology.MonitoredOptions.DeployDir,
			LogDir:    c.Topology.MonitoredOptions.LogDir,
			MonitorProcess: map[string]struct{}{
				"ts-store": {},
			},
			Config: map[string]any{
				"monitor.host":           s.Host,
				"monitor.error-log-path": s.LogDir,
				"monitor.disk-path":      s.DataDir,
				"monitor.aux-disk-path":  filepath.Join(s.DataDir, "wal"),
				"logging.path":           s.LogDir,
			},
		}
		instanceSet.Insert(s.Host)
		ins = append(ins, &TSMonitorInstance{
			Name: c.Name(),
			BaseInstance: BaseInstance{
				InstanceSpec: ms,
				Name:         c.Name(),
				Host:         s.Host,
				ManageHost:   s.ManageHost,
				ListenHost:   s.ListenHost,
				SSHP:         s.SSHPort,
				Source:       s.GetSource(),
				Dirs: []string{
					s.DeployDir,
					s.LogDir,
				},
			},
			topo: c.Topology,
		})
	}

	for _, s := range c.Topology.TSMetaServers {
		if instanceSet.Exist(s.Host) {
			for _, mo := range ins {
				if mo.GetHost() == s.Host {
					mo.(*TSMonitorInstance).BaseInstance.InstanceSpec.(*TSMonitorSpec).MonitorProcess["ts-meta"] = struct{}{}
				}
			}
			continue
		}
		ms := &TSMonitorSpec{
			OS:        s.OS,
			Arch:      s.Arch,
			DeployDir: c.Topology.MonitoredOptions.DeployDir,
			LogDir:    c.Topology.MonitoredOptions.LogDir,
			MonitorProcess: map[string]struct{}{
				"ts-meta": {},
			},
			Config: map[string]any{
				"monitor.host":           s.Host,
				"monitor.error-log-path": s.LogDir,
				"logging.path":           s.LogDir,
			},
		}
		instanceSet.Insert(s.Host)
		ins = append(ins, &TSMonitorInstance{
			Name: c.Name(),
			BaseInstance: BaseInstance{
				InstanceSpec: ms,
				Name:         c.Name(),
				Host:         s.Host,
				ManageHost:   s.ManageHost,
				ListenHost:   s.ListenHost,
				SSHP:         s.SSHPort,
				Source:       s.GetSource(),
				Dirs: []string{
					s.DeployDir,
					s.LogDir,
				},
			},
			topo: c.Topology,
		})
	}
	for _, s := range c.Topology.TSSqlServers {
		if instanceSet.Exist(s.Host) {
			for _, mo := range ins {
				if mo.GetHost() == s.Host {
					mo.(*TSMonitorInstance).BaseInstance.InstanceSpec.(*TSMonitorSpec).MonitorProcess["ts-sql"] = struct{}{}
				}
			}
			continue
		}
		ms := &TSMonitorSpec{
			OS:        s.OS,
			Arch:      s.Arch,
			DeployDir: c.Topology.MonitoredOptions.DeployDir,
			LogDir:    c.Topology.MonitoredOptions.LogDir,
			MonitorProcess: map[string]struct{}{
				"ts-sql": {},
			},
			Config: map[string]any{
				"monitor.host":           s.Host,
				"monitor.error-log-path": s.LogDir,
				"logging.path":           s.LogDir,
			},
		}
		instanceSet.Insert(s.Host)
		ins = append(ins, &TSMonitorInstance{
			Name: c.Name(),
			BaseInstance: BaseInstance{
				InstanceSpec: ms,
				Name:         c.Name(),
				Host:         s.Host,
				ManageHost:   s.ManageHost,
				ListenHost:   s.ListenHost,
				SSHP:         s.SSHPort,
				Source:       s.GetSource(),
				Dirs: []string{
					s.DeployDir,
					s.LogDir,
				},
			},
			topo: c.Topology,
		})
	}

	for _, s := range c.Topology.TSMonitorServers {
		if instanceSet.Exist(s.Host) {
			continue
		}
		s := s
		instanceSet.Insert(s.Host)
		ins = append(ins, &TSMonitorInstance{
			Name: c.Name(),
			BaseInstance: BaseInstance{
				InstanceSpec: s,
				Name:         c.Name(),
				Host:         s.Host,
				ManageHost:   s.ManageHost,
				ListenHost:   s.ListenHost,
				SSHP:         s.SSHPort,
				Source:       s.GetSource(),
				Dirs: []string{
					s.DeployDir,
					s.LogDir,
				},
			},
			topo: c.Topology,
		})
	}

	for _, mo := range ins {
		var procs []string
		for proc := range mo.(*TSMonitorInstance).BaseInstance.InstanceSpec.(*TSMonitorSpec).MonitorProcess {
			procs = append(procs, proc)
		}
		mo.(*TSMonitorInstance).BaseInstance.InstanceSpec.(*TSMonitorSpec).Config["monitor.process"] = strings.Join(procs, ",")
	}

	return ins
}

// TSMonitorInstance represent the ts-meta instance
type TSMonitorInstance struct {
	Name string
	BaseInstance
	topo *Specification
}

func (i *TSMonitorInstance) InitConfig(ctx context.Context, e ctxt.Executor, clusterName string, clusterVersion string, deployUser string, paths meta.DirPaths) error {
	topo := i.topo
	if err := i.BaseInstance.InitConfig(ctx, e, topo.GlobalOptions, deployUser, paths); err != nil {
		return err
	}
	spec := i.InstanceSpec.(*TSMonitorSpec)

	cfg := &scripts.TSMonitorScript{
		DeployDir: paths.Deploy,
		LogDir:    paths.Log,
	}

	fp := filepath.Join(paths.Cache, fmt.Sprintf("run_ts_monitor_%s_%d.sh", i.GetHost(), i.GetPort()))
	if err := cfg.ConfigToFile(fp); err != nil {
		return errors.WithStack(err)
	}

	dst := filepath.Join(paths.Deploy, "scripts", fmt.Sprintf("run_%s.sh", i.ComponentName())) // /path/to/scripts/run_ts-monitor.sh
	if err := e.Transfer(ctx, fp, dst, false, 0, false); err != nil {
		return errors.WithStack(err)
	}
	_, _, err := e.Execute(ctx, "chmod +x "+dst, false)
	if err != nil {
		return errors.WithStack(err)
	}

	globalConfig := topo.ServerConfigs.TsMonitor

	// set TLS configs
	//spec.Config, err = i.setTLSConfig(ctx, enableTLS, spec.Config, paths)
	//if err != nil {
	//	return errors.WithStack(err)
	//}

	configs := i.SetDefaultConfig(spec.Config, clusterName)

	if err = i.MergeServerConfig(ctx, e, globalConfig, configs, paths); err != nil {
		return errors.WithStack(err)
	}

	return checkConfig(ctx, e, i.ComponentName(), i.ComponentSource(), clusterVersion, i.OS(), i.Arch(), i.ComponentName()+".toml", paths, nil)
}

func (i *TSMonitorInstance) SetDefaultConfig(instanceConf map[string]any, clusterName string) map[string]any {
	if instanceConf == nil {
		instanceConf = make(map[string]any, 20)
	}

	if !i.topo.MonitoredOptions.TSMonitorEnabled {
		return instanceConf
	}

	// monitor
	for _, server := range i.topo.TSMetaServers {
		if i.Host == server.Host {
			instanceConf["monitor.metric-path"] = filepath.Join(server.LogDir, "metric")
			break
		}
	}
	for _, server := range i.topo.TSStoreServers {
		if i.Host == server.Host {
			instanceConf["monitor.metric-path"] = filepath.Join(server.LogDir, "metric")
			break
		}
	}
	for _, server := range i.topo.TSStoreServers {
		if i.Host == server.Host {
			instanceConf["monitor.metric-path"] = filepath.Join(server.LogDir, "metric")
			break
		}
	}

	instanceConf["report.database"] = strings.Replace(clusterName, "-", "_", -1)
	// TODO: report to monitor server address
	if len(i.topo.TSSqlServers) > 0 {
		instanceConf["report.address"] = fmt.Sprintf("%s:%d", i.topo.TSSqlServers[0].Host, i.topo.TSSqlServers[0].Port)
	}

	return instanceConf
}

// Ready implements Instance interface
func (i *TSMonitorInstance) Ready(ctx context.Context, e ctxt.Executor, timeout uint64, _ *tls.Config) error {
	return nil
}
