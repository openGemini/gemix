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
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/openGemini/gemix/pkg/cluster/ctxt"
	"github.com/openGemini/gemix/pkg/meta"
)

// Components names
const (
	ComponentTSMeta  = "ts-meta"
	ComponentTSSql   = "ts-sql"
	ComponentTSStore = "ts-store"
	ComponentGrafana = "grafana"
)

// Components sources
const (
	ComponentOpenGemini = "openGemini"
)

// Component represents a component of the cluster.
type Component interface {
	Name() string
	Role() string
	Instances() []Instance
}

// Instance represents the instance.
type Instance interface {
	InstanceSpec
	ID() string
	//Ready(context.Context, ctxt.Executor, uint64, *tls.Config) error
	InitConfig(ctx context.Context, e ctxt.Executor, clusterName string, clusterVersion string, deployUser string, paths meta.DirPaths) error
	//ScaleConfig(ctx context.Context, e ctxt.Executor, topo Topology, clusterName string, clusterVersion string, deployUser string, paths meta.DirPaths) error
	PrepareStart(ctx context.Context, tlsCfg *tls.Config) error
	ComponentName() string
	ComponentSource() string
	InstanceName() string
	ServiceName() string
	GetHost() string
	GetManageHost() string
	GetPort() int
	GetSSHPort() int
	DeployDir() string
	UsedPorts() []int
	UsedDirs() []string
	Status(ctx context.Context, timeout time.Duration, tlsCfg *tls.Config, pdList ...string) string
	Uptime(ctx context.Context, timeout time.Duration, tlsCfg *tls.Config) time.Duration
	DataDir() string
	LogDir() string
	OS() string // only linux supported now
	Arch() string
}

// BaseInstance implements some method of Instance interface..
type BaseInstance struct {
	InstanceSpec

	Name       string
	Host       string
	ManageHost string
	ListenHost string
	Port       int
	SSHP       int
	Source     string

	Ports    []int
	Dirs     []string
	StatusFn func(ctx context.Context, timeout time.Duration, tlsCfg *tls.Config, pdHosts ...string) string
	UptimeFn func(ctx context.Context, timeout time.Duration, tlsCfg *tls.Config) time.Duration
}

// Ready implements Instance interface
//func (i *BaseInstance) Ready(ctx context.Context, e ctxt.Executor, timeout uint64, _ *tls.Config) error {
//	return PortStarted(ctx, e, i.Port, timeout)
//}

// InitConfig init the service configuration.
//func (i *BaseInstance) InitConfig(ctx context.Context, e ctxt.Executor, opt GlobalOptions, user string, paths meta.DirPaths) (err error) {
//	return nil
//}

// MergeServerConfig merges the server configuration and overwrite the global configuration
func (i *BaseInstance) MergeServerConfig(ctx context.Context, e ctxt.Executor, globalConf, instanceConf map[string]any, paths meta.DirPaths) error {
	fp := filepath.Join(paths.Cache, fmt.Sprintf("%s-%s-%d.toml", i.ComponentName(), i.GetHost(), i.GetPort()))
	conf, err := Merge2Toml(i.ComponentName(), globalConf, instanceConf)
	if err != nil {
		return err
	}
	err = os.WriteFile(fp, conf, os.ModePerm)
	if err != nil {
		return err
	}
	dst := filepath.Join(paths.Deploy, "conf", fmt.Sprintf("%s.toml", i.ComponentName()))
	// transfer config
	return e.Transfer(ctx, fp, dst, false, 0, false)
}

// ID returns the identifier of this instance, the ID is constructed by host:port
func (i *BaseInstance) ID() string {
	return fmt.Sprintf("%s:%d", i.Host, i.Port)
}

// ComponentName implements Instance interface
func (i *BaseInstance) ComponentName() string {
	return i.Name
}

// ComponentSource implements Instance interface
func (i *BaseInstance) ComponentSource() string {
	if i.Source == "" {
		return i.Name
	}
	return i.Source
}

// InstanceName implements Instance interface
func (i *BaseInstance) InstanceName() string {
	if i.Port > 0 {
		return fmt.Sprintf("%s%d", i.Name, i.Port)
	}
	return i.ComponentName()
}

// ServiceName implements Instance interface
func (i *BaseInstance) ServiceName() string {
	name := i.Name
	if i.Port > 0 {
		return fmt.Sprintf("%s-%d.service", name, i.Port)
	}
	return fmt.Sprintf("%s.service", name)
}

// GetHost implements Instance interface
func (i *BaseInstance) GetHost() string {
	return i.Host
}

// GetManageHost implements Instance interface
func (i *BaseInstance) GetManageHost() string {
	if i.ManageHost != "" {
		return i.ManageHost
	}
	return i.Host
}

// GetListenHost implements Instance interface
func (i *BaseInstance) GetListenHost() string {
	if i.ListenHost == "" {
		// ipv6 address
		if strings.Contains(i.Host, ":") {
			return "::"
		}
		return "0.0.0.0"
	}
	return i.ListenHost
}

// GetSSHPort implements Instance interface
func (i *BaseInstance) GetSSHPort() int {
	return i.SSHP
}

// DeployDir implements Instance interface
func (i *BaseInstance) DeployDir() string {
	return reflect.Indirect(reflect.ValueOf(i.InstanceSpec)).FieldByName("DeployDir").String()
}

// TLSDir implements Instance interface
func (i *BaseInstance) TLSDir() string {
	return i.DeployDir()
}

// DataDir implements Instance interface
func (i *BaseInstance) DataDir() string {
	dataDir := reflect.Indirect(reflect.ValueOf(i.InstanceSpec)).FieldByName("DataDir")
	if !dataDir.IsValid() {
		return ""
	}

	// the default data_dir is relative to deploy_dir
	if dataDir.String() != "" && !strings.HasPrefix(dataDir.String(), "/") {
		return filepath.Join(i.DeployDir(), dataDir.String())
	}

	return dataDir.String()
}

// LogDir implements Instance interface
func (i *BaseInstance) LogDir() string {
	logDir := ""

	field := reflect.Indirect(reflect.ValueOf(i.InstanceSpec)).FieldByName("LogDir")
	if field.IsValid() {
		logDir = field.Interface().(string)
	}

	if logDir == "" {
		logDir = "log"
	}
	if !strings.HasPrefix(logDir, "/") {
		logDir = filepath.Join(i.DeployDir(), logDir)
	}
	return logDir
}

// OS implements Instance interface
func (i *BaseInstance) OS() string {
	v := reflect.Indirect(reflect.ValueOf(i.InstanceSpec)).FieldByName("OS")
	if !v.IsValid() {
		return ""
	}
	return v.Interface().(string)
}

// Arch implements Instance interface
func (i *BaseInstance) Arch() string {
	v := reflect.Indirect(reflect.ValueOf(i.InstanceSpec)).FieldByName("Arch")
	if !v.IsValid() {
		return ""
	}
	return v.Interface().(string)
}

// IsPatched implements Instance interface
func (i *BaseInstance) IsPatched() bool {
	v := reflect.Indirect(reflect.ValueOf(i.InstanceSpec)).FieldByName("Patched")
	if !v.IsValid() {
		return false
	}
	return v.Bool()
}

// SetPatched implements the Instance interface
func (i *BaseInstance) SetPatched(p bool) {
	v := reflect.Indirect(reflect.ValueOf(i.InstanceSpec)).FieldByName("Patched")
	if !v.CanSet() {
		return
	}
	v.SetBool(p)
}

// PrepareStart checks instance requirements before starting
func (i *BaseInstance) PrepareStart(ctx context.Context, tlsCfg *tls.Config) error {
	return nil
}

// GetPort implements Instance interface
func (i *BaseInstance) GetPort() int {
	return i.Port
}

// UsedPorts implements Instance interface
func (i *BaseInstance) UsedPorts() []int {
	return i.Ports
}

// UsedDirs implements Instance interface
func (i *BaseInstance) UsedDirs() []string {
	return i.Dirs
}

// Status implements Instance interface
func (i *BaseInstance) Status(ctx context.Context, timeout time.Duration, tlsCfg *tls.Config, pdList ...string) string {
	return i.StatusFn(ctx, timeout, tlsCfg, pdList...)
}

// Uptime implements Instance interface
func (i *BaseInstance) Uptime(ctx context.Context, timeout time.Duration, tlsCfg *tls.Config) time.Duration {
	return i.UptimeFn(ctx, timeout, tlsCfg)
}
