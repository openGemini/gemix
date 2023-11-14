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

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
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
