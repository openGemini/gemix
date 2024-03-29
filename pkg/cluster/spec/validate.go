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

// Copyright 2020 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package spec

import (
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/openGemini/gemix/pkg/gui"
	"github.com/openGemini/gemix/pkg/meta"
	"github.com/openGemini/gemix/pkg/set"
	"github.com/openGemini/gemix/pkg/utils"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// pre defined error types
var (
	errNSDeploy           = errNS.NewSubNamespace("deploy")
	errDeployDirConflict  = errNSDeploy.NewType("dir_conflict", utils.ErrTraitPreCheck)
	errDeployDirOverlap   = errNSDeploy.NewType("dir_overlap", utils.ErrTraitPreCheck)
	errDeployPortConflict = errNSDeploy.NewType("port_conflict", utils.ErrTraitPreCheck)
	ErrUserOrGroupInvalid = errors.New(`linux username and groupname must start with a lower case letter or an underscore, ` +
		`followed by lower case letters, digits, underscores, or dashes. ` +
		`Usernames may only be up to 32 characters long. ` +
		`Groupnames may only be up to 16 characters long.`)
)

// Linux username and groupname must start with a lower case letter or an underscore,
// followed by lower case letters, digits, underscores, or dashes.
// ref https://man7.org/linux/man-pages/man8/useradd.8.html
// ref https://man7.org/linux/man-pages/man8/groupadd.8.html
var (
	reUser  = regexp.MustCompile(`^[a-z_]([a-z0-9_-]{0,31}|[a-z0-9_-]{0,30}\$)$`)
	reGroup = regexp.MustCompile(`^[a-z_]([a-z0-9_-]{0,15})$`)
)

func fixDir(topo Topology) func(string) string {
	return func(dir string) string {
		if dir != "" {
			return Abs(topo.BaseTopo().GlobalOptions.User, dir)
		}
		return dir
	}
}

// DirAccessor stands for a directory accessor for an instance
type DirAccessor struct {
	dirKind  string
	accessor func(Instance, Topology) string
}

func dirAccessors() ([]DirAccessor, []DirAccessor) {
	instanceDirAccessor := []DirAccessor{
		{dirKind: "deploy directory", accessor: func(instance Instance, topo Topology) string { return instance.DeployDir() }},
		{dirKind: "data directory", accessor: func(instance Instance, topo Topology) string { return instance.DataDir() }},
		{dirKind: "log directory", accessor: func(instance Instance, topo Topology) string { return instance.LogDir() }},
	}
	hostDirAccessor := []DirAccessor{
		{dirKind: "monitor deploy directory", accessor: func(instance Instance, topo Topology) string {
			m := topo.BaseTopo().MonitoredOptions
			if m == nil {
				return ""
			}
			return m.DeployDir
		}},
		{dirKind: "monitor log directory", accessor: func(instance Instance, topo Topology) string {
			m := topo.BaseTopo().MonitoredOptions
			if m == nil {
				return ""
			}
			return m.LogDir
		}},
	}

	return instanceDirAccessor, hostDirAccessor
}

// DirEntry stands for a directory with attributes and instance
type DirEntry struct {
	clusterName string
	dirKind     string
	dir         string
	instance    Instance
}

func appendEntries(name string, topo Topology, inst Instance, dirAccessor DirAccessor, targets []DirEntry) []DirEntry {
	for _, dir := range strings.Split(fixDir(topo)(dirAccessor.accessor(inst, topo)), ",") {
		targets = append(targets, DirEntry{
			clusterName: name,
			dirKind:     dirAccessor.dirKind,
			dir:         dir,
			instance:    inst,
		})
	}

	return targets
}

// CheckClusterDirConflict checks cluster dir conflict or overlap
func CheckClusterDirConflict(clusterList map[string]Metadata, clusterName string, topo Topology) error {
	instanceDirAccessor, hostDirAccessor := dirAccessors()
	var currentEntries []DirEntry
	var existingEntries []DirEntry

	// rebuild existing disk status
	for name, metadata := range clusterList {
		if name == clusterName {
			continue
		}

		gotTopo := metadata.GetTopology()

		gotTopo.IterInstance(func(inst Instance) {
			for _, dirAccessor := range instanceDirAccessor {
				existingEntries = appendEntries(name, gotTopo, inst, dirAccessor, existingEntries)
			}
		})
		IterHost(gotTopo, func(inst Instance) {
			for _, dirAccessor := range hostDirAccessor {
				existingEntries = appendEntries(name, gotTopo, inst, dirAccessor, existingEntries)
			}
		})
	}

	topo.IterInstance(func(inst Instance) {
		for _, dirAccessor := range instanceDirAccessor {
			currentEntries = appendEntries(clusterName, topo, inst, dirAccessor, currentEntries)
		}
	})
	IterHost(topo, func(inst Instance) {
		for _, dirAccessor := range hostDirAccessor {
			currentEntries = appendEntries(clusterName, topo, inst, dirAccessor, currentEntries)
		}
	})

	for _, d1 := range currentEntries {
		// data_dir is relative to deploy_dir by default, so they can be with
		// same (sub) paths as long as the deploy_dirs are different
		if d1.dirKind == "data directory" && !strings.HasPrefix(d1.dir, "/") {
			continue
		}
		for _, d2 := range existingEntries {
			if d1.instance.GetManageHost() != d2.instance.GetManageHost() {
				continue
			}

			//// ignore conflict in the case when both sides are monitor and either one of them
			//// is marked as ignore exporter.
			//if strings.HasPrefix(d1.dirKind, "monitor") &&
			//	strings.HasPrefix(d2.dirKind, "monitor") &&
			//	(d1.instance.IgnoreMonitorAgent() || d2.instance.IgnoreMonitorAgent()) {
			//	continue
			//}

			if d1.dir == d2.dir && d1.dir != "" {
				properties := map[string]string{
					"ThisDirKind":    d1.dirKind,
					"ThisDir":        d1.dir,
					"ThisComponent":  d1.instance.ComponentName(),
					"ThisHost":       d1.instance.GetManageHost(),
					"ExistCluster":   d2.clusterName,
					"ExistDirKind":   d2.dirKind,
					"ExistDir":       d2.dir,
					"ExistComponent": d2.instance.ComponentName(),
					"ExistHost":      d2.instance.GetManageHost(),
				}
				zap.L().Info("Meet deploy directory conflict", zap.Any("info", properties))
				return errDeployDirConflict.New("Deploy directory conflicts to an existing cluster").WithProperty(gui.SuggestionFromTemplate(`
The directory you specified in the topology file is:
  Directory: {{ColorKeyword}}{{.ThisDirKind}} {{.ThisDir}}{{ColorReset}}
  Component: {{ColorKeyword}}{{.ThisComponent}} {{.ThisHost}}{{ColorReset}}

It conflicts to a directory in the existing cluster:
  Existing Cluster Name: {{ColorKeyword}}{{.ExistCluster}}{{ColorReset}}
  Existing Directory:    {{ColorKeyword}}{{.ExistDirKind}} {{.ExistDir}}{{ColorReset}}
  Existing Component:    {{ColorKeyword}}{{.ExistComponent}} {{.ExistHost}}{{ColorReset}}

Please change to use another directory or another host.
`, properties))
			}
		}
	}

	return CheckClusterDirOverlap(currentEntries)
}

// CheckClusterDirOverlap checks cluster dir overlaps with data or log.
// this should only be used across clusters.
// we don't allow to deploy log under data, and vise versa.
func CheckClusterDirOverlap(entries []DirEntry) error {
	ignore := func(d1, d2 DirEntry) bool {
		return (d1.instance.GetManageHost() != d2.instance.GetManageHost()) ||
			d1.dir == "" || d2.dir == "" ||
			strings.HasSuffix(d1.dirKind, "deploy directory") ||
			strings.HasSuffix(d2.dirKind, "deploy directory")
	}
	for i := 0; i < len(entries)-1; i++ {
		d1 := entries[i]
		for j := i + 1; j < len(entries); j++ {
			d2 := entries[j]
			if ignore(d1, d2) {
				continue
			}

			if utils.IsSubDir(d1.dir, d2.dir) || utils.IsSubDir(d2.dir, d1.dir) {
				properties := map[string]string{
					"ThisDirKind":   d1.dirKind,
					"ThisDir":       d1.dir,
					"ThisComponent": d1.instance.ComponentName(),
					"ThisHost":      d1.instance.GetManageHost(),
					"ThatDirKind":   d2.dirKind,
					"ThatDir":       d2.dir,
					"ThatComponent": d2.instance.ComponentName(),
					"ThatHost":      d2.instance.GetManageHost(),
				}
				zap.L().Info("Meet deploy directory overlap", zap.Any("info", properties))
				return errDeployDirOverlap.New("Deploy directory overlaps to another instance").WithProperty(gui.SuggestionFromTemplate(`
The directory you specified in the topology file is:
  Directory: {{ColorKeyword}}{{.ThisDirKind}} {{.ThisDir}}{{ColorReset}}
  Component: {{ColorKeyword}}{{.ThisComponent}} {{.ThisHost}}{{ColorReset}}

It overlaps to another instance:
  Other Directory: {{ColorKeyword}}{{.ThatDirKind}} {{.ThatDir}}{{ColorReset}}
  Other Component: {{ColorKeyword}}{{.ThatComponent}} {{.ThatHost}}{{ColorReset}}

Please modify the topology file and try again.
`, properties))
			}
		}
	}

	return nil
}

// CheckClusterPortConflict checks cluster port conflict
func CheckClusterPortConflict(clusterList map[string]Metadata, clusterName string, topo Topology) error {
	type Entry struct {
		clusterName   string
		componentName string
		port          int
		instance      Instance
	}

	currentEntries := []Entry{}
	existingEntries := []Entry{}

	for name, metadata := range clusterList {
		if name == clusterName {
			continue
		}

		metadata.GetTopology().IterInstance(func(inst Instance) {
			for _, port := range inst.UsedPorts() {
				existingEntries = append(existingEntries, Entry{
					clusterName:   name,
					componentName: inst.ComponentName(),
					port:          port,
					instance:      inst,
				})
			}
		})
	}

	topo.IterInstance(func(inst Instance) {
		for _, port := range inst.UsedPorts() {
			currentEntries = append(currentEntries, Entry{
				componentName: inst.ComponentName(),
				port:          port,
				instance:      inst,
			})
		}
	})

	for _, p1 := range currentEntries {
		for _, p2 := range existingEntries {
			if p1.instance.GetManageHost() != p2.instance.GetManageHost() {
				continue
			}

			if p1.port == p2.port {
				// build the conflict info
				properties := map[string]string{
					"ThisPort":       strconv.Itoa(p1.port),
					"ThisComponent":  p1.componentName,
					"ThisHost":       p1.instance.GetManageHost(),
					"ExistCluster":   p2.clusterName,
					"ExistPort":      strconv.Itoa(p2.port),
					"ExistComponent": p2.componentName,
					"ExistHost":      p2.instance.GetManageHost(),
				}

				// build error message
				zap.L().Info("Meet deploy port conflict", zap.Any("info", properties))
				return errDeployPortConflict.New("Deploy port conflicts to an existing cluster").WithProperty(gui.SuggestionFromTemplate(`
The port you specified in the topology file is:
  Port:      {{ColorKeyword}}{{.ThisPort}}{{ColorReset}}
  Component: {{ColorKeyword}}{{.ThisComponent}} {{.ThisHost}}{{ColorReset}}

It conflicts to a port in the existing cluster:
  Existing Cluster Name: {{ColorKeyword}}{{.ExistCluster}}{{ColorReset}}
  Existing Port:         {{ColorKeyword}}{{.ExistPort}}{{ColorReset}}
  Existing Component:    {{ColorKeyword}}{{.ExistComponent}} {{.ExistHost}}{{ColorReset}}

Please change to use another port or another host.
`, properties))
			}
		}
	}

	return nil
}

//lint:ignore U1000 keep this
func getHostFromAddress(addr string) string {
	host, _ := utils.ParseHostPort(addr)
	return host
}

// platformConflictsDetect checks for conflicts in topology for different OS / Arch
// set to the same host / IP
func (s *Specification) platformConflictsDetect() error {
	type (
		conflict struct {
			os   string
			arch string
			cfg  string
		}
	)

	platformStats := map[string]conflict{}
	topoSpec := reflect.ValueOf(s).Elem()
	topoType := reflect.TypeOf(s).Elem()

	for i := 0; i < topoSpec.NumField(); i++ {
		if isSkipField(topoSpec.Field(i)) {
			continue
		}

		compSpecs := topoSpec.Field(i)
		for index := 0; index < compSpecs.Len(); index++ {
			compSpec := reflect.Indirect(compSpecs.Index(index))
			// check hostname
			host := compSpec.FieldByName("Host").String()
			cfg := strings.Split(topoType.Field(i).Tag.Get("yaml"), ",")[0] // without meta
			if host == "" {
				return errors.Errorf("`%s` contains empty host field", cfg)
			}

			// platform conflicts
			stat := conflict{
				cfg: cfg,
			}
			if j, found := findField(compSpec, "OS"); found {
				stat.os = compSpec.Field(j).String()
			}
			if j, found := findField(compSpec, "Arch"); found {
				stat.arch = compSpec.Field(j).String()
			}

			prev, exist := platformStats[host]
			if exist {
				if prev.os != stat.os || prev.arch != stat.arch {
					return &meta.ValidateErr{
						Type:   meta.TypeMismatch,
						Target: "platform",
						LHS:    fmt.Sprintf("%s:%s/%s", prev.cfg, prev.os, prev.arch),
						RHS:    fmt.Sprintf("%s:%s/%s", stat.cfg, stat.os, stat.arch),
						Value:  host,
					}
				}
			}
			platformStats[host] = stat
		}
	}
	return nil
}

func (s *Specification) portInvalidDetect() error {
	topoSpec := reflect.ValueOf(s).Elem()
	topoType := reflect.TypeOf(s).Elem()

	checkPort := func(idx int, compSpec reflect.Value) error {
		compSpec = reflect.Indirect(compSpec)
		cfg := strings.Split(topoType.Field(idx).Tag.Get("yaml"), ",")[0]

		for i := 0; i < compSpec.NumField(); i++ {
			if strings.HasSuffix(compSpec.Type().Field(i).Name, "Port") {
				port := int(compSpec.Field(i).Int())
				// for NgPort, 0 means default and -1 means disable
				if compSpec.Type().Field(i).Name == "NgPort" && (port == -1 || port == 0) {
					continue
				}
				if port < 1 || port > 65535 {
					portField := strings.Split(compSpec.Type().Field(i).Tag.Get("yaml"), ",")[0]
					return errors.Errorf("`%s` of %s=%d is invalid, port should be in the range [1, 65535]", cfg, portField, port)
				}
			}
		}
		return nil
	}

	for i := 0; i < topoSpec.NumField(); i++ {
		compSpecs := topoSpec.Field(i)

		// check on struct
		if compSpecs.Kind() == reflect.Struct {
			if err := checkPort(i, compSpecs); err != nil {
				return err
			}
			continue
		}

		// check on slice
		for index := 0; index < compSpecs.Len(); index++ {
			compSpec := reflect.Indirect(compSpecs.Index(index))
			if err := checkPort(i, compSpec); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Specification) portConflictsDetect() error {
	type (
		usedPort struct {
			host string
			port int
		}
		conflict struct {
			tp  string
			cfg string
		}
	)

	portTypes := []string{
		"Port",
		"StatusPort",
		"PeerPort",
		"ClientPort",
		"WebPort",
		"TCPPort",
		"HTTPPort",
		"FlashServicePort",
		"FlashProxyPort",
		"FlashProxyStatusPort",
		"ClusterPort",
		"NgPort",
	}

	portStats := map[usedPort]conflict{}
	uniqueHosts := set.NewStringSet()
	topoSpec := reflect.ValueOf(s).Elem()
	topoType := reflect.TypeOf(s).Elem()

	for i := 0; i < topoSpec.NumField(); i++ {
		if isSkipField(topoSpec.Field(i)) {
			continue
		}

		compSpecs := topoSpec.Field(i)
		for index := 0; index < compSpecs.Len(); index++ {
			compSpec := reflect.Indirect(compSpecs.Index(index))

			// check hostname
			host := compSpec.FieldByName("Host").String()
			cfg := strings.Split(topoType.Field(i).Tag.Get("yaml"), ",")[0] // without meta
			if host == "" {
				return errors.Errorf("`%s` contains empty host field", cfg)
			}
			uniqueHosts.Insert(host)

			// Ports conflicts
			for _, portType := range portTypes {
				if j, found := findField(compSpec, portType); found {
					item := usedPort{
						host: host,
						port: int(compSpec.Field(j).Int()),
					}
					tp := compSpec.Type().Field(j).Tag.Get("yaml")
					prev, exist := portStats[item]
					if exist {
						return &meta.ValidateErr{
							Type:   meta.TypeConflict,
							Target: "port",
							LHS:    fmt.Sprintf("%s:%s.%s", prev.cfg, item.host, prev.tp),
							RHS:    fmt.Sprintf("%s:%s.%s", cfg, item.host, tp),
							Value:  item.port,
						}
					}
					portStats[item] = conflict{
						tp:  tp,
						cfg: cfg,
					}
				}
			}
		}
	}

	return nil
}

func (s *Specification) dirConflictsDetect() error {
	type (
		usedDir struct {
			host string
			dir  string
		}
		conflict struct {
			tp       string
			cfg      string
			imported bool
		}
	)

	dirTypes := []string{
		"DataDir",
		"DeployDir",
	}

	// usedInfo => type
	var dirStats = map[usedDir]conflict{}

	topoSpec := reflect.ValueOf(s).Elem()
	topoType := reflect.TypeOf(s).Elem()

	for i := 0; i < topoSpec.NumField(); i++ {
		if isSkipField(topoSpec.Field(i)) {
			continue
		}

		compSpecs := topoSpec.Field(i)
		for index := 0; index < compSpecs.Len(); index++ {
			compSpec := reflect.Indirect(compSpecs.Index(index))
			// check hostname
			host := compSpec.FieldByName("Host").String()
			cfg := strings.Split(topoType.Field(i).Tag.Get("yaml"), ",")[0] // without meta
			if host == "" {
				return errors.Errorf("`%s` contains empty host field", cfg)
			}

			// Directory conflicts
			for _, dirType := range dirTypes {
				j, found := findField(compSpec, dirType)
				if !found {
					continue
				}

				// `yaml:"data_dir,omitempty"`
				tp := strings.Split(compSpec.Type().Field(j).Tag.Get("yaml"), ",")[0]
				for _, dir := range strings.Split(compSpec.Field(j).String(), ",") {
					dir = strings.TrimSpace(dir)
					item := usedDir{
						host: host,
						dir:  dir,
					}
					// data_dir is relative to deploy_dir by default, so they can be with
					// same (sub) paths as long as the deploy_dirs are different
					if item.dir != "" && !strings.HasPrefix(item.dir, "/") {
						continue
					}
					prev, exist := dirStats[item]
					// not checking between imported nodes
					if exist {
						return &meta.ValidateErr{
							Type:   meta.TypeConflict,
							Target: "directory",
							LHS:    fmt.Sprintf("%s:%s.%s", prev.cfg, item.host, prev.tp),
							RHS:    fmt.Sprintf("%s:%s.%s", cfg, item.host, tp),
							Value:  item.dir,
						}
					}
					// not reporting error for nodes imported from openGemini-Ansible, but keep
					// their dirs in the map to check if other nodes are using them
					dirStats[item] = conflict{
						tp:       tp,
						cfg:      cfg,
						imported: false,
					}
				}
			}
		}
	}

	return nil
}

// CountDir counts for dir paths used by any instance in the cluster with the same
// prefix, useful to find potential path conflicts
func (s *Specification) CountDir(targetHost, dirPrefix string) int {
	dirTypes := []string{
		"DeployDir",
		"DataDir",
		"LogDir",
	}

	// path -> count
	dirStats := make(map[string]int)
	count := 0
	topoSpec := reflect.ValueOf(s).Elem()
	dirPrefix = Abs(s.GlobalOptions.User, dirPrefix)

	addHostDir := func(host, deployDir, dir string) {
		if !strings.HasPrefix(dir, "/") {
			dir = filepath.Join(deployDir, dir)
		}
		dir = Abs(s.GlobalOptions.User, dir)
		dirStats[host+dir]++
	}

	for i := 0; i < topoSpec.NumField(); i++ {
		if isSkipField(topoSpec.Field(i)) {
			continue
		}

		compSpecs := topoSpec.Field(i)
		for index := 0; index < compSpecs.Len(); index++ {
			compSpec := reflect.Indirect(compSpecs.Index(index))
			deployDir := compSpec.FieldByName("DeployDir").String()
			host := compSpec.FieldByName("Host").String()
			if compSpec.FieldByName("ManageHost").String() != "" {
				host = compSpec.FieldByName("ManageHost").String()
			}

			for _, dirType := range dirTypes {
				j, found := findField(compSpec, dirType)
				if !found {
					continue
				}

				dir := compSpec.Field(j).String()

				switch dirType { // the same as in instance.go for (*instance)
				case "DeployDir":
					addHostDir(host, deployDir, "")
				case "DataDir":
					// the default data_dir is relative to deploy_dir
					if dir == "" {
						addHostDir(host, deployDir, dir)
						continue
					}
					for _, dataDir := range strings.Split(dir, ",") {
						dataDir = strings.TrimSpace(dataDir)
						if dataDir != "" {
							addHostDir(host, deployDir, dataDir)
						}
					}
				case "LogDir":
					field := compSpec.FieldByName("LogDir")
					if field.IsValid() {
						dir = field.Interface().(string)
					}

					if dir == "" {
						dir = "log"
					}
					addHostDir(host, deployDir, strings.TrimSpace(dir))
				}
			}
		}
	}

	for k, v := range dirStats {
		if k == targetHost+dirPrefix || strings.HasPrefix(k, targetHost+dirPrefix+"/") {
			count += v
		}
	}

	return count
}

func (s *Specification) validateTLSEnabled() error {
	if !s.GlobalOptions.TLSEnabled {
		return nil
	}

	// check for component with no tls support
	compList := make([]Component, 0)
	s.IterComponent(func(c Component) {
		if len(c.Instances()) > 0 {
			compList = append(compList, c)
		}
	})

	for _, c := range compList {
		switch c.Name() {
		case ComponentOpenGemini,
			ComponentGrafana:
		default:
			return errors.Errorf("component %s is not supported in TLS enabled cluster", c.Name())
		}
	}
	return nil
}

func (s *Specification) validateUserGroup() error {
	gOpts := s.GlobalOptions
	if user := gOpts.User; !reUser.MatchString(user) {
		return errors.WithMessagef(ErrUserOrGroupInvalid, "`global` of user='%s' is invalid", user)
	}
	// if group is nil, then we'll set it to the same as user
	if group := gOpts.Group; group != "" && !reGroup.MatchString(group) {
		return errors.WithMessagef(ErrUserOrGroupInvalid, "`global` of group='%s' is invalid", group)
	}
	return nil
}

func (s *Specification) validateTSMetaNames() error {
	// check ts-meta-server name
	metaNames := set.NewStringSet()
	for _, m := range s.TSMetaServers {
		if m.Name == "" {
			continue
		}

		if metaNames.Exist(m.Name) {
			return errors.Errorf("component ts_meta_servers.name is not supported duplicated, the name %s is duplicated", m.Name)
		}
		metaNames.Insert(m.Name)
	}
	return nil
}

// validateMonitorAgent checks for conflicts in topology for different ignore_exporter
// settings for multiple instances on the same host / IP
//
//lint:ignore U1000 keep this
func (s *Specification) validateMonitorAgent() error {
	type (
		conflict struct {
			ignore bool
			cfg    string
		}
	)
	agentStats := map[string]conflict{}
	topoSpec := reflect.ValueOf(s).Elem()
	topoType := reflect.TypeOf(s).Elem()

	for i := 0; i < topoSpec.NumField(); i++ {
		if isSkipField(topoSpec.Field(i)) {
			continue
		}

		compSpecs := topoSpec.Field(i)
		for index := 0; index < compSpecs.Len(); index++ {
			compSpec := reflect.Indirect(compSpecs.Index(index))

			// check hostname
			host := compSpec.FieldByName("Host").String()
			cfg := strings.Split(topoType.Field(i).Tag.Get("yaml"), ",")[0] // without meta
			if host == "" {
				return errors.Errorf("`%s` contains empty host field", cfg)
			}

			// agent conflicts
			stat := conflict{}
			if j, found := findField(compSpec, "IgnoreExporter"); found {
				stat.ignore = compSpec.Field(j).Bool()
				stat.cfg = cfg
			}

			prev, exist := agentStats[host]
			if exist {
				if prev.ignore != stat.ignore {
					return &meta.ValidateErr{
						Type:   meta.TypeMismatch,
						Target: "ignore_exporter",
						LHS:    fmt.Sprintf("%s:%v", prev.cfg, prev.ignore),
						RHS:    fmt.Sprintf("%s:%v", stat.cfg, stat.ignore),
						Value:  host,
					}
				}
			}
			agentStats[host] = stat
		}
	}
	return nil
}

// Validate validates the topology specification and produce error if the
// specification invalid (e.g: port conflicts or directory conflicts)
func (s *Specification) Validate() error {
	validators := []func() error{
		s.validateTLSEnabled,
		s.platformConflictsDetect,
		s.portInvalidDetect,
		s.portConflictsDetect,
		s.dirConflictsDetect,
		s.validateUserGroup,
		s.validateTSMetaNames,
	}

	for _, v := range validators {
		if err := v(); err != nil {
			return errors.WithStack(err)
		}
	}

	return RelativePathDetect(s, isSkipField)
}

// RelativePathDetect detect if some specific path is relative path and report error
func RelativePathDetect(topo any, isSkipField func(reflect.Value) bool) error {
	pathTypes := []string{
		"ConfigFilePath",
		"RuleDir",
		"DashboardDir",
	}

	topoSpec := reflect.ValueOf(topo).Elem()

	for i := 0; i < topoSpec.NumField(); i++ {
		if isSkipField(topoSpec.Field(i)) {
			continue
		}

		compSpecs := topoSpec.Field(i)
		for index := 0; index < compSpecs.Len(); index++ {
			compSpec := reflect.Indirect(compSpecs.Index(index))

			// Relateve path detect
			for _, field := range pathTypes {
				if j, found := findField(compSpec, field); found {
					// `yaml:"xxxx,omitempty"`
					fieldName := strings.Split(compSpec.Type().Field(j).Tag.Get("yaml"), ",")[0]
					localPath := compSpec.Field(j).String()
					if localPath != "" && !strings.HasPrefix(localPath, "/") {
						return fmt.Errorf("relative path is not allowed for field %s: %s", fieldName, localPath)
					}
				}
			}
		}
	}

	return nil
}
