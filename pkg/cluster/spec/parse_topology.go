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
	"path"
	"path/filepath"
	"reflect"
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
	//	return nil, errors.New("missing required options for yaml configuration file")
	//}

	// TODO: Update with default values
	//updateWithGlobalDefaults(&yamlSpec)

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

func expandRelativePath(user string, topo any) {
	v := reflect.Indirect(reflect.ValueOf(topo).Elem())

	switch v.Kind() {
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			ref := reflect.New(v.Index(i).Type())
			ref.Elem().Set(v.Index(i))
			expandRelativePath(user, ref.Interface())
			v.Index(i).Set(ref.Elem())
		}
	case reflect.Struct:
		// We should deal with DeployDir first, because DataDir and LogDir depends on it
		dirs := []string{"DeployDir", "DataDir", "LogDir"}
		for _, dir := range dirs {
			f := v.FieldByName(dir)
			if !f.IsValid() || f.String() == "" {
				continue
			}
			switch dir {
			case "DeployDir":
				f.SetString(Abs(user, f.String()))
			case "DataDir":
				// Some components supports multiple data dirs split by comma
				ds := strings.Split(f.String(), ",")
				ads := []string{}
				for _, d := range ds {
					if strings.HasPrefix(d, "/") {
						ads = append(ads, d)
					} else {
						ads = append(ads, path.Join(v.FieldByName("DeployDir").String(), d))
					}
				}
				f.SetString(strings.Join(ads, ","))
			case "LogDir":
				if !strings.HasPrefix(f.String(), "/") {
					f.SetString(path.Join(v.FieldByName("DeployDir").String(), f.String()))
				}
			}
		}
		// Deal with all fields (expandRelativePath will do nothing on string filed)
		for i := 0; i < v.NumField(); i++ {
			// We don't deal with GlobalOptions because relative path in GlobalOptions.Data has special meaning
			if v.Type().Field(i).Name == "GlobalOptions" {
				continue
			}
			ref := reflect.New(v.Field(i).Type())
			ref.Elem().Set(v.Field(i))
			expandRelativePath(user, ref.Interface())
			v.Field(i).Set(ref.Elem())
		}
	case reflect.Ptr:
		expandRelativePath(user, v.Interface())
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
