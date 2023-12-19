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
	"os"
	"path/filepath"
	"strings"

	"github.com/joomcode/errorx"
	"github.com/openGemini/gemix/pkg/utils"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

var (
	defaultDeployUser = "gemini"
	errNSTopolohy     = errorx.NewNamespace("topology")
	// ErrTopologyReadFailed is ErrTopologyReadFailed
	ErrTopologyReadFailed = errNSTopolohy.NewType("read_failed", utils.ErrTraitPreCheck)
	// ErrTopologyParseFailed is ErrTopologyParseFailed
	ErrTopologyParseFailed = errNSTopolohy.NewType("parse_failed", utils.ErrTraitPreCheck)
)

// ReadYamlFile read yaml content from file
func ReadYamlFile(file string) ([]byte, error) {
	yamlFile, err := os.ReadFile(file)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to read topology file %s\n. Please check whether your topology file %s exists and try again.\n", file, file)
	}
	return yamlFile, nil
}

func ReadFromYaml(file string) (*Specification, error) {
	var err error
	yamlFile, err := ReadYamlFile(file)
	if err != nil {
		return nil, err
	}
	var myTopo map[string]any
	if err = yaml.Unmarshal(yamlFile, &myTopo); err != nil {
		return nil, err
	}
	yamlFile, _ = yaml.Marshal(myTopo)

	var yamlSpec = &Specification{}
	if err = yaml.UnmarshalStrict(yamlFile, yamlSpec); err != nil {
		return nil, errors.Wrapf(err, "Failed to parse topology file %s\n. Please check the syntax of your topology file %s and try again.\n", file, file)
	}

	zap.L().Debug("Parse topology file succeeded", zap.Any("topology", yamlSpec))

	// TODO: check required options
	//if pass := checkRequiredOptions(yamlSpec); !pass {
	//	return nil, errors.New("missing requitred options for yaml configuration file")
	//}

	// TODO: Update with default values
	//updataWithGlobalDefaults(&yamlSpec)

	return yamlSpec, nil
}

// ParseTopologyYaml read yaml content from `file` and unmarshal it to `out`
// ignoreGlobal ignore global variables in file, only ignoreGlobal with a index of 0 is effective
func ParseTopologyYaml(file string, out Topology, ignoreGlobal ...bool) error {
	zap.L().Debug("Parse topology file", zap.String("file", file))

	yamlFile, err := ReadYamlFile(file)
	if err != nil {
		return err
	}

	// keep the global config in out
	if len(ignoreGlobal) > 0 && ignoreGlobal[0] {
		var newTopo map[string]any
		if err := yaml.Unmarshal(yamlFile, &newTopo); err != nil {
			return err
		}
		for k := range newTopo {
			switch k {
			case "global",
				"server_configs":
				delete(newTopo, k)
			}
		}
		yamlFile, _ = yaml.Marshal(newTopo)
	}

	if err = yaml.UnmarshalStrict(yamlFile, out); err != nil {
		return errors.WithMessagef(err, "failed to parse topology file %s\n. Please check the syntax of your topology file %s and try again", file, file)
	}

	zap.L().Debug("Parse topology file succeeded", zap.Any("topology", out))
	return nil
}

// Abs returns the absolute path
func Abs(user, path string) string {
	// trim whitespaces before joining
	user = strings.TrimSpace(user)
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "~/") {
		if user == "root" {
			path = filepath.Join("/root", path[2:])
		} else {
			path = filepath.Join("/home", user, path[2:])
		}
	} else if !strings.HasPrefix(path, "/") {
		path = filepath.Join("/home", user, path)
	}
	return filepath.Clean(path)
}

// ExpandRelativeDir fill DeployDir, DataDir and LogDir to absolute path
func ExpandRelativeDir(topo Topology) {
	expandRelativePath(deployUser(topo), topo)
}

func expandRelativePath(user string, topology Topology) {
	topo := topology.(*Specification)
	topo.GlobalOptions.DeployDir = Abs(user, topo.GlobalOptions.DeployDir)
	topo.GlobalOptions.LogDir = Abs(user, topo.GlobalOptions.LogDir)

	for i := range topo.TSMetaServers {
		server := topo.TSMetaServers[i]
		server.DeployDir = Abs(user, server.DeployDir)
		server.LogDir = Abs(user, server.LogDir)
		server.DataDir = Abs(user, server.DataDir)
	}

	for i := range topo.TSSqlServers {
		server := topo.TSSqlServers[i]
		server.DeployDir = Abs(user, server.DeployDir)
		server.LogDir = Abs(user, server.LogDir)
	}

	for i := range topo.TSStoreServers {
		server := topo.TSStoreServers[i]
		server.DeployDir = Abs(user, server.DeployDir)
		server.LogDir = Abs(user, server.LogDir)
		server.DataDir = Abs(user, server.DataDir)
	}

	for i := range topo.TSMonitorServers {
		server := topo.TSMonitorServers[i]
		server.DeployDir = Abs(user, server.DeployDir)
		server.LogDir = Abs(user, server.LogDir)
	}

	for i := range topo.Grafanas {
		server := topo.Grafanas[i]
		server.DeployDir = Abs(user, server.DeployDir)
	}
}

func deployUser(topo Topology) string {
	base := topo.BaseTopo()
	if base.GlobalOptions == nil || base.GlobalOptions.User == "" {
		return defaultDeployUser
	}
	return base.GlobalOptions.User
}

// SetValueFromGlobal set the default value from global
func SetValueFromGlobal(topo *Specification) {
	setValueFromGlobal(topo)
}

func setValueFromGlobal(topo *Specification) {

}
