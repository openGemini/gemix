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
	"github.com/openGemini/gemix/utils"
	"github.com/pkg/errors"
)

// TSStoreSpec represents the ts-store topology specification in topology.yaml
type TSStoreSpec struct {
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
	DataDir   string `yaml:"data_dir"`
	MetaDir   string `yaml:"meta_dir"`

	// port specification
	IngestPort int `yaml:"ingest_port" default:"8400"`
	SelectPort int `yaml:"select_port" default:"8401"`
	GossipPort int `yaml:"gossip_port" default:"8011"`

	Config map[string]any `yaml:"config,omitempty" validate:"config:ignore"`
}

func (s *TSStoreSpec) SSH() (string, int) {
	host := s.Host
	if s.ManageHost != "" {
		host = s.ManageHost
	}
	return host, s.SSHPort
}

// Role returns the component role of the instance
func (s *TSStoreSpec) Role() string {
	return ComponentTSStore
}

// GetManageHost returns the manage host of the instance
func (s *TSStoreSpec) GetManageHost() string {
	if s.ManageHost != "" {
		return s.ManageHost
	}
	return s.Host
}

// GetSource returns source to download the component
func (s *TSStoreSpec) GetSource() string {
	if s.Source == "" {
		return ComponentOpenGemini
	}
	return s.Source
}

// TSStoreComponent represents ts-meta component.
type TSStoreComponent struct{ Topology *Specification }

// Name implements Component interface.
func (c *TSStoreComponent) Name() string {
	return ComponentTSStore
}

// Role implements Component interface.
func (c *TSStoreComponent) Role() string {
	return ComponentTSStore
}

// Instances implements Component interface.
func (c *TSStoreComponent) Instances() []Instance {
	ins := make([]Instance, 0, len(c.Topology.TSStoreServers))
	for _, s := range c.Topology.TSStoreServers {
		s := s
		ins = append(ins, &TSStoreInstance{
			Name: c.Name(),
			BaseInstance: BaseInstance{
				InstanceSpec: s,
				Name:         c.Name(),
				Host:         s.Host,
				ManageHost:   s.ManageHost,
				ListenHost:   s.ListenHost,
				Port:         s.SelectPort, // do not change me
				SSHP:         s.SSHPort,
				Source:       s.GetSource(),

				Ports: []int{
					s.IngestPort,
					s.SelectPort,
					s.GossipPort,
				},
				Dirs: []string{
					s.DeployDir,
					s.LogDir,
					s.DataDir,
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

// TSStoreInstance represent the ts-meta instance
type TSStoreInstance struct {
	Name string
	BaseInstance
	topo *Specification
}

func (i *TSStoreInstance) InitConfig(ctx context.Context, e ctxt.Executor, clusterName string, clusterVersion string, deployUser string, paths meta.DirPaths) error {
	topo := i.topo
	if err := i.BaseInstance.InitConfig(ctx, e, topo.GlobalOptions, deployUser, paths); err != nil {
		return err
	}

	//enableTLS := topo.GlobalOptions.TLSEnabled
	spec := i.InstanceSpec.(*TSStoreSpec)

	cfg := &scripts.TSStoreScript{
		DeployDir: paths.Deploy,
		LogDir:    paths.Log,
	}

	fp := filepath.Join(paths.Cache, fmt.Sprintf("run_ts_store_%s_%d.sh", i.GetHost(), i.GetPort()))
	if err := cfg.ConfigToFile(fp); err != nil {
		return errors.WithStack(err)
	}
	dst := filepath.Join(paths.Deploy, "scripts", "run_ts_store.sh")

	if err := e.Transfer(ctx, fp, dst, false, 0, false); err != nil {
		return errors.WithStack(err)
	}

	_, _, err := e.Execute(ctx, "chmod +x "+dst, false)
	if err != nil {
		return errors.WithStack(err)
	}

	globalConfig := topo.ServerConfigs.TsStore

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

func (i *TSStoreInstance) SetDefaultConfig(instanceConf map[string]any) map[string]any {
	if instanceConf == nil {
		instanceConf = make(map[string]any, 20)
	}
	var metaPeerAddrs []string
	for _, metaSpec := range i.topo.TSMetaServers {
		metaPeerAddrs = append(metaPeerAddrs, utils.JoinHostPort(metaSpec.Host, metaSpec.PeerPort))
	}
	instanceConf["common.meta-join"] = metaPeerAddrs

	var tsStoreSpec *TSStoreSpec
	for _, storeSpec := range i.topo.TSStoreServers {
		if i.Host == storeSpec.Host && i.Port == storeSpec.SelectPort {
			tsStoreSpec = storeSpec
		}
	}
	instanceConf["data.store-ingest-addr"] = utils.JoinHostPort(i.Host, tsStoreSpec.IngestPort)
	instanceConf["data.store-select-addr"] = utils.JoinHostPort(i.Host, tsStoreSpec.SelectPort)
	instanceConf["data.store-data-dir"] = tsStoreSpec.DataDir
	instanceConf["data.store-wal-dir"] = tsStoreSpec.DataDir
	//instanceConf["data.store-meta-dir"] = tsStoreSpec.DataDir
	instanceConf["logging.path"] = tsStoreSpec.LogDir

	instanceConf["gossip.bind-address"] = i.Host
	instanceConf["gossip.store-bind-port"] = tsStoreSpec.GossipPort
	instanceConf["gossip.meta-bind-port"] = tsStoreSpec.GossipPort - 1 // just for ts-store compatibility

	var metaGossipAddrs []string
	for _, metaSpec := range i.topo.TSMetaServers {
		metaGossipAddrs = append(metaGossipAddrs, utils.JoinHostPort(metaSpec.Host, metaSpec.GossipPort))
	}
	instanceConf["gossip.members"] = metaGossipAddrs

	return instanceConf
}
