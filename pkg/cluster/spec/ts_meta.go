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

// TSMetaSpec represents the ts-meta topology specification in topology.yaml
type TSMetaSpec struct {
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
	ClientPort int `yaml:"client_port" default:"8091"`
	PeerPort   int `yaml:"peer_port" default:"8092"`
	RaftPort   int `yaml:"raft_port"  default:"8088"`
	GossipPort int `yaml:"gossip_port"  default:"8010"`

	Config map[string]any `yaml:"config,omitempty" validate:"config:ignore"`
}

// Status queries current status of the instance
func (s *TSMetaSpec) Status(ctx context.Context, timeout time.Duration, tlsCfg *tls.Config, _ ...string) string {
	//if timeout < time.Second {
	//	timeout = statusQueryTimeout
	//}

	//addr := utils.JoinHostPort(s.GetManageHost(), s.ClientPort)
	//pc := api.NewPDClient(ctx, []string{addr}, timeout, tlsCfg)
	//
	//// check health
	//err := pc.CheckHealth()
	//if err != nil {
	//	return "Down"
	//}
	//
	//// find leader node
	//leader, err := pc.GetLeader()
	//if err != nil {
	//	return "ERR"
	//}
	res := "Up"
	//if s.Name == leader.Name {
	//	res += "|L"
	//}
	return res
}

func (s *TSMetaSpec) SSH() (string, int) {
	host := s.Host
	if s.ManageHost != "" {
		host = s.ManageHost
	}
	return host, s.SSHPort
}

// Role returns the component role of the instance
func (s *TSMetaSpec) Role() string {
	return ComponentTSMeta
}

// GetManageHost returns the manage host of the instance
func (s *TSMetaSpec) GetManageHost() string {
	if s.ManageHost != "" {
		return s.ManageHost
	}
	return s.Host
}

// GetSource returns source to download the component
func (s *TSMetaSpec) GetSource() string {
	if s.Source == "" {
		return ComponentOpenGemini
	}
	return s.Source
}

// TSMetaComponent represents ts-meta component.
type TSMetaComponent struct{ Topology *Specification }

// Name implements Component interface.
func (c *TSMetaComponent) Name() string {
	return ComponentTSMeta
}

// Role implements Component interface.
func (c *TSMetaComponent) Role() string {
	return ComponentTSMeta
}

// Instances implements Component interface.
func (c *TSMetaComponent) Instances() []Instance {
	ins := make([]Instance, 0, len(c.Topology.TSMetaServers))
	for _, s := range c.Topology.TSMetaServers {
		s := s
		ins = append(ins, &TSMetaInstance{
			Name: c.Name(),
			BaseInstance: BaseInstance{
				InstanceSpec: s,
				Name:         c.Name(),
				Host:         s.Host,
				ManageHost:   s.ManageHost,
				ListenHost:   utils.Ternary(s.ListenHost != "", s.ListenHost, c.Topology.BaseTopo().GlobalOptions.ListenHost).(string),
				Port:         s.ClientPort,
				SSHP:         s.SSHPort,
				Source:       s.GetSource(),

				Ports: []int{
					s.ClientPort,
					s.PeerPort,
					s.RaftPort,
					s.GossipPort,
				},
				Dirs: []string{
					s.DeployDir,
					s.LogDir,
					s.DataDir,
				},
				StatusFn: s.Status,
				UptimeFn: func(_ context.Context, timeout time.Duration, tlsCfg *tls.Config) time.Duration {
					return UptimeByHost(s.GetManageHost(), s.ClientPort, timeout, tlsCfg)
				},
			},
			topo: c.Topology,
		})
	}
	return ins
}

// TSMetaInstance represent the ts-meta instance
type TSMetaInstance struct {
	Name string
	BaseInstance
	topo *Specification
}

func (i *TSMetaInstance) InitConfig(ctx context.Context, e ctxt.Executor, clusterName string, clusterVersion string, deployUser string, paths meta.DirPaths) error {
	topo := i.topo
	if err := i.BaseInstance.InitConfig(ctx, e, topo.GlobalOptions, deployUser, paths); err != nil {
		return err
	}
	spec := i.InstanceSpec.(*TSMetaSpec)

	cfg := &scripts.TSMetaScript{
		DeployDir: paths.Deploy,
		LogDir:    paths.Log,
	}

	fp := filepath.Join(paths.Cache, fmt.Sprintf("run_ts_meta_%s_%d.sh", i.GetHost(), i.GetPort()))
	if err := cfg.ConfigToFile(fp); err != nil {
		return errors.WithStack(err)
	}
	dst := filepath.Join(paths.Deploy, "scripts", "run_ts-meta.sh")
	if err := e.Transfer(ctx, fp, dst, false, 0, false); err != nil {
		return errors.WithStack(err)
	}
	_, _, err := e.Execute(ctx, "chmod +x "+dst, false)
	if err != nil {
		return errors.WithStack(err)
	}

	globalConfig := topo.ServerConfigs.TsMeta

	// set TLS configs
	//spec.Config, err = i.setTLSConfig(ctx, enableTLS, spec.Config, paths)
	//if err != nil {
	//	return err
	//}

	configs := i.SetDefaultConfig(spec.Config)

	if err = i.MergeServerConfig(ctx, e, globalConfig, configs, paths); err != nil {
		return errors.WithStack(err)
	}

	return checkConfig(ctx, e, i.ComponentName(), i.ComponentSource(), clusterVersion, i.OS(), i.Arch(), i.ComponentName()+".toml", paths, nil)
}

func (i *TSMetaInstance) SetDefaultConfig(instanceConf map[string]any) map[string]any {
	if instanceConf == nil {
		instanceConf = make(map[string]any, 20)
	}
	var metaPeerAddrs []string
	for _, metaSpec := range i.topo.TSMetaServers {
		metaPeerAddrs = append(metaPeerAddrs, utils.JoinHostPort(metaSpec.Host, metaSpec.PeerPort))
	}
	var tsMetaSpec = i.InstanceSpec.(*TSMetaSpec)
	instanceConf["common.meta-join"] = metaPeerAddrs
	instanceConf["meta.bind-address"] = utils.JoinHostPort(i.Host, tsMetaSpec.RaftPort)
	instanceConf["meta.http-bind-address"] = utils.JoinHostPort(i.Host, tsMetaSpec.ClientPort)
	instanceConf["meta.rpc-bind-address"] = utils.JoinHostPort(i.Host, tsMetaSpec.PeerPort)
	instanceConf["meta.dir"] = tsMetaSpec.DataDir
	instanceConf["logging.path"] = tsMetaSpec.LogDir

	instanceConf["gossip.bind-address"] = i.Host
	instanceConf["gossip.meta-bind-port"] = tsMetaSpec.GossipPort
	instanceConf["gossip.store-bind-port"] = tsMetaSpec.GossipPort + 1 // just for ts-meta compatibility

	var metaGossipAddrs []string
	for _, metaSpec := range i.topo.TSMetaServers {
		metaGossipAddrs = append(metaGossipAddrs, utils.JoinHostPort(metaSpec.Host, metaSpec.GossipPort))
	}
	instanceConf["gossip.members"] = metaGossipAddrs

	// monitor
	if i.topo.MonitoredOptions.TSMonitorEnabled {
		instanceConf["monitor.pushers"] = "file"
		instanceConf["monitor.store-enabled"] = true
		instanceConf["monitor.store-path"] = filepath.Join(tsMetaSpec.LogDir, "metric", "meta-metric.data")
	}
	return instanceConf
}
