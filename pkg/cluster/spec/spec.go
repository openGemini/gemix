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
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/creasty/defaults"
	"github.com/openGemini/gemix/pkg/meta"
	"github.com/openGemini/gemix/pkg/utils"
	"github.com/pkg/errors"
)

// FullHostType is the type of fullhost operations
type FullHostType string

const (
	// FullArchType cpu-arch type
	FullArchType FullHostType = "Arch"
	// FullOSType kernel-name
	FullOSType FullHostType = "OS"
)

type (
	// InstanceSpec represent a instance specification
	InstanceSpec interface {
		Role() string
		SSH() (string, int)
	}

	// GlobalOptions represents the global options for all groups in topology
	// specification in topology.yaml
	GlobalOptions struct {
		User            string               `yaml:"user,omitempty" default:"gemini"`
		Group           string               `yaml:"group,omitempty"`
		SSHPort         int                  `yaml:"ssh_port,omitempty" default:"22" validate:"ssh_port:editable"`
		TLSEnabled      bool                 `yaml:"enable_tls,omitempty"`
		DeployDir       string               `yaml:"deploy_dir,omitempty" default:"deploy"`
		LogDir          string               `yaml:"log_dir,omitempty" default:"logs"`
		DataDir         string               `yaml:"data_dir,omitempty" default:"data"`
		ResourceControl meta.ResourceControl `yaml:"resource_control,omitempty" validate:"resource_control:editable"`
		OS              string               `yaml:"os,omitempty" default:"linux"`
		Arch            string               `yaml:"arch,omitempty" default:"amd64"`
		//Custom          any                  `yaml:"custom,omitempty" validate:"custom:ignore"`
	}

	// TSMonitoredOptions represents the monitored configuration
	TSMonitoredOptions struct {
		TSMonitorEnabled bool   `yaml:"ts_monitor_enabled,omitempty" default:"false"`
		DeployDir        string `yaml:"deploy_dir,omitempty"`
		LogDir           string `yaml:"log_dir,omitempty"`
	}

	// ServerConfigs represents the server runtime configuration
	ServerConfigs struct {
		TsMeta    map[string]any    `yaml:"ts-meta"`
		TsSql     map[string]any    `yaml:"ts-sql"`
		TsStore   map[string]any    `yaml:"ts-store"`
		TsMonitor map[string]any    `yaml:"ts-monitor"`
		Grafana   map[string]string `yaml:"grafana"`
	}

	// Specification represents the specification of topology.yaml
	Specification struct {
		GlobalOptions    GlobalOptions      `yaml:"global,omitempty" validate:"global:editable"`
		MonitoredOptions TSMonitoredOptions `yaml:"monitored,omitempty" validate:"monitored:editable"`
		ServerConfigs    ServerConfigs      `yaml:"server_configs,omitempty" validate:"server_configs:ignore"`
		TSMetaServers    []*TSMetaSpec      `yaml:"ts_meta_servers"`
		TSSqlServers     []*TSSqlSpec       `yaml:"ts_sql_servers"`
		TSStoreServers   []*TSStoreSpec     `yaml:"ts_store_servers"`
		TSMonitorServers []*TSMonitorSpec   `yaml:"ts_monitor_servers,omitempty"`
		Grafanas         []*GrafanaSpec     `yaml:"grafana_servers,omitempty"`
		//DashboardServers []*DashboardSpec `yaml:"opengemini_dashboard_servers,omitempty"`
		//Monitors         []*PrometheusSpec    `yaml:"monitoring_servers"`
	}
)

// Topology represents specification of the cluster.
type Topology interface {
	BaseTopo() *BaseTopo
	// Validate validates the topology specification and produce error if the
	// specification invalid (e.g: port conflicts or directory conflicts)
	Validate() error

	ComponentsByStartOrder() []Component
	ComponentsByStopOrder() []Component
	//ComponentsByUpdateOrder(curVer string) []Component
	IterInstance(fn func(instance Instance), concurrency ...int)
	CountDir(host string, dir string) int // count how many time a path is used by instances in cluster
	//TLSConfig(dir string) (*tls.Config, error)
	//Merge(that Topology) Topology // TODO: for update
	FillHostArchOrOS(hostArchmap map[string]string, fullType FullHostType) error
	GetGrafanaConfig() map[string]string
}

type BaseTopo struct {
	GlobalOptions *GlobalOptions
	MasterList    []string

	Grafanas []*GrafanaSpec
}

// BaseMeta is the base info of metadata.
type BaseMeta struct {
	User    string
	Group   string
	Version string
}

// Metadata of a cluster.
type Metadata interface {
	GetTopology() Topology
	SetTopology(topo Topology)
	GetBaseMeta() *BaseMeta
}

// AllComponentNames contains the names of all components.
// should include all components in ComponentsByStartOrder
func AllComponentNames() (roles []string) {
	tp := &Specification{}
	tp.IterComponent(func(c Component) {
		roles = append(roles, c.Name())
	})
	return
}

// UnmarshalYAML implements the yaml.Unmarshaler interface,
// it sets the default values when unmarshaling the topology file
func (s *Specification) UnmarshalYAML(unmarshal func(any) error) error {
	type topology Specification
	if err := unmarshal((*topology)(s)); err != nil {
		return errors.WithStack(err)
	}

	// set default values from tag
	if err := defaults.Set(s); err != nil {
		return errors.WithStack(err)
	}

	// populate custom default values as needed
	if err := fillCustomDefaults(&s.GlobalOptions, s); err != nil {
		return err
	}

	// TODO: validate yaml configs
	//return s.Validate()
	return nil
}

// fillDefaults tries to fill custom fields to their default values
func fillCustomDefaults(globalOptions *GlobalOptions, data any) error {
	v := reflect.ValueOf(data).Elem()
	t := v.Type()

	var err error
	for i := 0; i < t.NumField(); i++ {
		if err = setCustomDefaults(globalOptions, v.Field(i)); err != nil {
			return err
		}
	}

	return nil
}

var (
	globalOptionTypeName  = reflect.TypeOf(GlobalOptions{}).Name()
	monitorOptionTypeName = reflect.TypeOf(TSMonitoredOptions{}).Name()
	serverConfigsTypeName = reflect.TypeOf(ServerConfigs{}).Name()
)

// Skip global/monitored options
func isSkipField(field reflect.Value) bool {
	tp := field.Type().Name()
	return tp == globalOptionTypeName || tp == monitorOptionTypeName || tp == serverConfigsTypeName
}

func setDefaultDir(parent, role, port string, field reflect.Value) {
	if field.String() != "" {
		return
	}
	if defaults.CanUpdate(field.Interface()) {
		dir := fmt.Sprintf("%s-%s", role, port)
		field.Set(reflect.ValueOf(filepath.Join(parent, dir)))
	}
}

func getPort(v reflect.Value) string {
	for i := 0; i < v.NumField(); i++ {
		switch v.Type().Field(i).Name {
		case "Port", "ClientPort":
			return fmt.Sprintf("%d", v.Field(i).Int())
		}
	}
	return ""
}

func setCustomDefaults(globalOptions *GlobalOptions, field reflect.Value) error {
	if !field.CanSet() || isSkipField(field) {
		return nil
	}

	switch field.Kind() {
	case reflect.Slice:
		for i := 0; i < field.Len(); i++ {
			if err := setCustomDefaults(globalOptions, field.Index(i)); err != nil {
				return err
			}
		}
	case reflect.Struct:
		ref := reflect.New(field.Type())
		ref.Elem().Set(field)
		if err := fillCustomDefaults(globalOptions, ref.Interface()); err != nil {
			return err
		}
		field.Set(ref.Elem())
	case reflect.Ptr:
		if err := setCustomDefaults(globalOptions, field.Elem()); err != nil {
			return err
		}
	}

	if field.Kind() != reflect.Struct {
		return nil
	}

	for j := 0; j < field.NumField(); j++ {
		switch field.Type().Field(j).Name {
		case "SSHPort":
			if field.Field(j).Int() != 0 {
				continue
			}
			field.Field(j).Set(reflect.ValueOf(globalOptions.SSHPort))
		case "DataDir":
			if imported := reflect.Indirect(field).FieldByName("Imported"); imported.IsValid() && imported.Interface().(bool) {
				setDefaultDir(globalOptions.DataDir, field.Addr().Interface().(InstanceSpec).Role(), getPort(field), field.Field(j))
			}

			dataDir := field.Field(j).String()

			// If the per-instance data_dir already have a value, skip filling default values
			// and ignore any value in global data_dir, the default values are filled only
			// when the pre-instance data_dir is empty
			if dataDir != "" {
				continue
			}
			// If the data dir in global options is an absolute path, append current
			// value to the global and has a comp-port sub directory
			if strings.HasPrefix(globalOptions.DataDir, "/") {
				field.Field(j).Set(reflect.ValueOf(filepath.Join(
					globalOptions.DataDir,
					fmt.Sprintf("%s-%s", field.Addr().Interface().(InstanceSpec).Role(), getPort(field)),
				)))
				continue
			}
			// If the data dir in global options is empty or a relative path, keep it be relative
			// Our run_*.sh start scripts are run inside deploy_path, so the final location
			// will be deploy_path/global.data_dir
			// (the default value of global.data_dir is "data")
			if globalOptions.DataDir == "" {
				field.Field(j).Set(reflect.ValueOf("data"))
			} else {
				field.Field(j).Set(reflect.ValueOf(globalOptions.DataDir))
			}
		case "DeployDir":
			setDefaultDir(globalOptions.DeployDir, field.Addr().Interface().(InstanceSpec).Role(), getPort(field), field.Field(j))
		case "LogDir":
			if imported := reflect.Indirect(field).FieldByName("Imported"); imported.IsValid() && imported.Interface().(bool) {
				setDefaultDir(globalOptions.LogDir, field.Addr().Interface().(InstanceSpec).Role(), getPort(field), field.Field(j))
			}

			logDir := field.Field(j).String()

			// If the per-instance log_dir already have a value, skip filling default values
			// and ignore any value in global log_dir, the default values are filled only
			// when the pre-instance log_dir is empty
			if logDir != "" {
				continue
			}
			// If the log dir in global options is an absolute path, append current
			// value to the global and has a comp-port sub directory
			if strings.HasPrefix(globalOptions.LogDir, "/") {
				field.Field(j).Set(reflect.ValueOf(filepath.Join(
					globalOptions.LogDir,
					fmt.Sprintf("%s-%s", field.Addr().Interface().(InstanceSpec).Role(), getPort(field)),
				)))
				continue
			}
			// If the log dir in global options is empty or a relative path, keep it be relative
			// Our run_*.sh start scripts are run inside deploy_path, so the final location
			// will be deploy_path/global.log_dir
			// (the default value of global.log_dir is "log")
			if globalOptions.LogDir == "" {
				field.Field(j).Set(reflect.ValueOf("log"))
			} else {
				field.Field(j).Set(reflect.ValueOf(globalOptions.LogDir))
			}
		case "Arch":
			switch strings.ToLower(field.Field(j).String()) {
			// replace "x86_64" with amd64, they are the same in our repo
			case "x86_64":
				field.Field(j).Set(reflect.ValueOf("amd64"))
			// replace "aarch64" with arm64
			case "aarch64":
				field.Field(j).Set(reflect.ValueOf("arm64"))
			}

			// convert to lower case
			if field.Field(j).String() != "" {
				field.Field(j).Set(reflect.ValueOf(strings.ToLower(field.Field(j).String())))
			}
		case "OS":
			// convert to lower case
			if field.Field(j).String() != "" {
				field.Field(j).Set(reflect.ValueOf(strings.ToLower(field.Field(j).String())))
			}
		}
	}

	return nil
}

// GetTSMetaListWithManageHost returns a list of ts-meta API hosts of the current cluster
func (s *Specification) GetTSMetaListWithManageHost() []string {
	var tsMetaList []string
	for _, meta := range s.TSMetaServers {
		tsMetaList = append(tsMetaList, utils.JoinHostPort(meta.GetManageHost(), meta.ClientPort))
	}

	return tsMetaList
}

// BaseTopo implements Specification interface.
func (s *Specification) BaseTopo() *BaseTopo {
	return &BaseTopo{
		GlobalOptions: &s.GlobalOptions,
		MasterList:    s.GetTSMetaListWithManageHost(),
		Grafanas:      s.Grafanas,
	}
}

// ComponentsByStopOrder return component in the order need to stop.
func (s *Specification) ComponentsByStopOrder() (comps []Component) {
	comps = s.ComponentsByStartOrder()
	// revert order
	i := 0
	j := len(comps) - 1
	for i < j {
		comps[i], comps[j] = comps[j], comps[i]
		i++
		j--
	}
	return
}

// ComponentsByStartOrder return component in the order need to start.
func (s *Specification) ComponentsByStartOrder() (comps []Component) {
	// "ts-meta", "ts-store", "ts-sql", "ts-data", "ts-monitor", "grafana"
	comps = append(comps, &TSMetaComponent{Topology: s})
	comps = append(comps, &TSStoreComponent{Topology: s})
	comps = append(comps, &TSSqlComponent{Topology: s})
	//comps = append(comps, &TSDataComponent{s})
	comps = append(comps, &TSMonitorComponent{Topology: s})
	comps = append(comps, &GrafanaComponent{Topology: s})
	return
}

// IterComponent iterates all components in component starting order
func (s *Specification) IterComponent(fn func(comp Component)) {
	for _, comp := range s.ComponentsByStartOrder() {
		fn(comp)
	}
}

// IterInstance iterates all instances in component starting order
func (s *Specification) IterInstance(fn func(instance Instance), concurrency ...int) {
	maxWorkers := 1
	wg := sync.WaitGroup{}
	if len(concurrency) > 0 && concurrency[0] > 1 {
		maxWorkers = concurrency[0]
	}
	workerPool := make(chan struct{}, maxWorkers)

	for _, comp := range s.ComponentsByStartOrder() {
		for _, inst := range comp.Instances() {
			wg.Add(1)
			workerPool <- struct{}{}
			go func(inst Instance) {
				defer func() {
					<-workerPool
					wg.Done()
				}()
				fn(inst)
			}(inst)
		}
	}
	wg.Wait()
}

// FillHostArchOrOS fills the topology with the given host->arch
func (s *Specification) FillHostArchOrOS(hostArchOrOS map[string]string, fullType FullHostType) error {
	if err := FillHostArchOrOS(s, hostArchOrOS, fullType); err != nil {
		return err
	}

	return nil
}

// FillHostArchOrOS fills the topology with the given host->arch
func FillHostArchOrOS(s *Specification, hostArchOrOS map[string]string, fullType FullHostType) error {
	for host, arch := range hostArchOrOS {
		switch arch {
		case "x86_64":
			hostArchOrOS[host] = "amd64"
		case "aarch64":
			hostArchOrOS[host] = "arm64"
		default:
			hostArchOrOS[host] = strings.ToLower(arch)
		}
	}

	if fullType == FullOSType {
		for _, server := range s.TSMetaServers {
			if server.OS == "" {
				server.OS = hostArchOrOS[server.Host]
			}
		}
		for _, server := range s.TSSqlServers {
			if server.OS == "" {
				server.OS = hostArchOrOS[server.Host]
			}
		}
		for _, server := range s.TSStoreServers {
			if server.OS == "" {
				server.OS = hostArchOrOS[server.Host]
			}
		}
		for _, server := range s.TSMonitorServers {
			if server.OS == "" {
				server.OS = hostArchOrOS[server.Host]
			}
		}
		for _, server := range s.Grafanas {
			if server.OS == "" {
				server.OS = hostArchOrOS[server.Host]
			}
		}
	}
	if fullType == FullArchType {
		for _, server := range s.TSMetaServers {
			if server.Arch == "" {
				server.Arch = hostArchOrOS[server.Host]
			}
		}
		for _, server := range s.TSSqlServers {
			if server.Arch == "" {
				server.Arch = hostArchOrOS[server.Host]
			}
		}
		for _, server := range s.TSStoreServers {
			if server.Arch == "" {
				server.Arch = hostArchOrOS[server.Host]
			}
		}
		for _, server := range s.TSMonitorServers {
			if server.Arch == "" {
				server.Arch = hostArchOrOS[server.Host]
			}
		}
		for _, server := range s.Grafanas {
			if server.Arch == "" {
				server.Arch = hostArchOrOS[server.Host]
			}
		}
	}
	return nil
}

// PackagePath return the openGemini component tar.gz path
func PackagePath(comp string, version string, os string, arch string) string {
	fileName := fmt.Sprintf("%s-%s-%s-%s.tar.gz", comp, version, os, arch)
	return ProfilePath(OpenGeminiPackageCacheDir, fileName)
}

// PackageGrafanaPath return the grafana component tar.gz path
func PackageGrafanaPath(comp string, version string, os string, arch string) string {
	fileName := fmt.Sprintf("%s-%s.%s-%s.tar.gz", comp, version, os, arch)
	return ProfilePath(OpenGeminiPackageCacheDir, fileName)
}

// GetGrafanaConfig returns global grafana configurations
func (s *Specification) GetGrafanaConfig() map[string]string {
	return s.ServerConfigs.Grafana
}
