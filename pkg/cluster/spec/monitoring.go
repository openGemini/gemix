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
	"time"

	"github.com/openGemini/gemix/pkg/cluster/ctxt"
	"github.com/openGemini/gemix/pkg/cluster/template/scripts"
	"github.com/openGemini/gemix/pkg/meta"
	"github.com/openGemini/gemix/pkg/utils"
	"github.com/pkg/errors"
)

// TSServerSpec represents the ts-server topology specification in topology.yaml
type TSServerSpec struct {
	Arch string `yaml:"arch,omitempty"`
	OS   string `yaml:"os,omitempty"`

	Source string `yaml:"source,omitempty" validate:"source:editable"`

	// Use Name to get the name with a default value if it's empty.
	Name string `yaml:"name"`

	Host       string `yaml:"host"`
	ManageHost string `yaml:"manage_host,omitempty" validate:"manage_host:editable"`
	ListenHost string `yaml:"listen_host,omitempty"`
	SSHPort    int    `yaml:"ssh_port,omitempty" validate:"ssh_port:editable"`

	LogDir    string `yaml:"log_dir,omitempty"`
	DeployDir string `yaml:"deploy_dir,omitempty"`
	DataDir   string `yaml:"data_dir,omitempty"`

	// port specification
	Port       int `yaml:"port" default:"8186"`
	RaftPort   int `yaml:"raft_port"  default:"8188"`
	ClientPort int `yaml:"client_port" default:"8191"`
	PeerPort   int `yaml:"peer_port" default:"8192"`
	IngestPort int `yaml:"ingest_port" default:"8410"`
	SelectPort int `yaml:"select_port" default:"8411"`

	Config map[string]any `yaml:"config,omitempty" validate:"config:ignore"`
}

func (s *TSServerSpec) SSH() (string, int) {
	host := s.Host
	if s.ManageHost != "" {
		host = s.ManageHost
	}
	return host, s.SSHPort
}

// Role returns the component role of the instance
func (s *TSServerSpec) Role() string {
	return ComponentTSServer
}

// GetManageHost returns the manage host of the instance
func (s *TSServerSpec) GetManageHost() string {
	if s.ManageHost != "" {
		return s.ManageHost
	}
	return s.Host
}

// GetSource returns source to download the component
func (s *TSServerSpec) GetSource() string {
	if s.Source == "" {
		return ComponentOpenGemini
	}
	return s.Source
}

// TSServerComponent represents ts-meta component.
type TSServerComponent struct{ Topology }

// Name implements Component interface.
func (c *TSServerComponent) Name() string {
	return ComponentTSServer
}

// Role implements Component interface.
func (c *TSServerComponent) Role() string {
	return ComponentTSServer
}

// Instances implements Component interface.
func (c *TSServerComponent) Instances() []Instance {
	servers := c.BaseTopo().Monitors
	ins := make([]Instance, 0, len(servers))

	for _, s := range servers {
		s := s
		ins = append(ins, &TSServerInstance{
			BaseInstance: BaseInstance{
				InstanceSpec: s,
				Name:         c.Name(),
				Host:         s.Host,
				ManageHost:   s.ManageHost,
				ListenHost:   c.Topology.BaseTopo().GlobalOptions.ListenHost,
				Port:         s.Port,
				SSHP:         s.SSHPort,
				Source:       s.GetSource(),

				Ports: []int{
					s.Port,
					s.RaftPort,
					s.ClientPort,
					s.PeerPort,
					s.IngestPort,
					s.SelectPort,
				},
				Dirs: []string{
					s.DeployDir,
					s.LogDir,
					s.DataDir,
				},
				StatusFn: func(_ context.Context, timeout time.Duration, _ *tls.Config, _ ...string) string {
					return statusByHost(s.GetManageHost(), s.Port, "/status", timeout, nil)
				},
				UptimeFn: func(_ context.Context, timeout time.Duration, tlsCfg *tls.Config) time.Duration {
					return UptimeByHost(s.GetManageHost(), s.Port, timeout, tlsCfg)
				},
			},
			topo: c.Topology,
		})
	}
	return ins
}

// TSServerInstance represent the ts-meta instance
type TSServerInstance struct {
	BaseInstance
	topo Topology
}

func (i *TSServerInstance) InitConfig(ctx context.Context, e ctxt.Executor, clusterName string, clusterVersion string, deployUser string, paths meta.DirPaths) error {
	//topo := i.topo.(*Specification)
	gOpts := *i.topo.BaseTopo().GlobalOptions
	if err := i.BaseInstance.InitConfig(ctx, e, gOpts, deployUser, paths); err != nil {
		return err
	}

	//enableTLS := topo.GlobalOptions.TLSEnabled
	spec := i.InstanceSpec.(*TSServerSpec)

	cfg := &scripts.TSServerScript{
		DeployDir: paths.Deploy,
		LogDir:    paths.Log,
	}

	fp := filepath.Join(paths.Cache, fmt.Sprintf("run_ts_server_%s_%d.sh", i.GetHost(), i.GetPort()))
	if err := cfg.ConfigToFile(fp); err != nil {
		return errors.WithStack(err)
	}
	dst := filepath.Join(paths.Deploy, "scripts", "run_ts-server.sh")

	if err := e.Transfer(ctx, fp, dst, false, 0, false); err != nil {
		return errors.WithStack(err)
	}

	_, _, err := e.Execute(ctx, "chmod +x "+dst, false)
	if err != nil {
		return errors.WithStack(err)
	}

	//globalConfig := topo.ServerConfigs.TsStore

	// set TLS configs
	//spec.Config, err = i.setTLSConfig(ctx, enableTLS, spec.Config, paths)
	//if err != nil {
	//	return err
	//}

	configs := i.SetDefaultConfig(spec.Config)

	if err = i.MergeServerConfig(ctx, e, nil, configs, paths); err != nil {
		return errors.WithStack(err)
	}

	return checkConfig(ctx, e, i.ComponentName(), i.ComponentSource(), clusterVersion, i.OS(), i.Arch(), i.ComponentName()+".toml", paths, nil)
}

func (i *TSServerInstance) SetDefaultConfig(instanceConf map[string]any) map[string]any {
	if instanceConf == nil {
		instanceConf = make(map[string]any, 20)
	}
	var serverSpec = i.InstanceSpec.(*TSServerSpec)

	instanceConf["common.meta-join"] = []string{utils.JoinHostPort(serverSpec.Host, serverSpec.PeerPort)}
	instanceConf["meta.bind-address"] = utils.JoinHostPort(i.Host, serverSpec.RaftPort)
	instanceConf["meta.http-bind-address"] = utils.JoinHostPort(i.Host, serverSpec.ClientPort)
	instanceConf["meta.rpc-bind-address"] = utils.JoinHostPort(i.Host, serverSpec.PeerPort)
	instanceConf["meta.dir"] = filepath.Join(serverSpec.DataDir, "meta")
	instanceConf["data.store-ingest-addr"] = utils.JoinHostPort(i.Host, serverSpec.IngestPort)
	instanceConf["data.store-select-addr"] = utils.JoinHostPort(i.Host, serverSpec.SelectPort)
	instanceConf["data.store-data-dir"] = serverSpec.DataDir
	instanceConf["data.store-wal-dir"] = serverSpec.DataDir
	instanceConf["http.bind-address"] = utils.JoinHostPort(i.Host, serverSpec.Port)
	instanceConf["logging.path"] = serverSpec.LogDir

	// monitor
	if i.topo.(*Specification).MonitoredOptions.TSMonitorEnabled {
		instanceConf["monitor.pushers"] = "file"
		instanceConf["monitor.store-enabled"] = false //TODO
		instanceConf["monitor.store-path"] = filepath.Join(serverSpec.LogDir, "metric", "server-metric.data")
	}
	return instanceConf
}
