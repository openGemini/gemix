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

package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type StartOptions struct {
	User           string
	SkipCreateUser bool
}

type ClusterOptions struct {
	Name     string
	Version  string
	User     string
	Key      string
	Password string
	SshType  SSHType
	YamlPath string
}

func SaveClusterOptionsToFile(filename string, cluster ClusterOptions) error {
	data, err := json.Marshal(cluster)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, data, 0750)
	return err
}

func LoadClusterOptionsFromFile(filename string) (ClusterOptions, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return ClusterOptions{}, err
	}

	var cluster ClusterOptions
	err = json.Unmarshal(data, &cluster)
	if err != nil {
		return ClusterOptions{}, err
	}

	return cluster, nil
}

func CheckClusterNameValid(name string) bool {
	files, err := ioutil.ReadDir(ClusterInfoDir)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return false
	}

	for _, file := range files {
		if file.Name() == name {
			return false
		}
	}
	return true
}

func CheckClusterNameExist(name string) bool {
	files, err := ioutil.ReadDir(ClusterInfoDir)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return false
	}

	for _, file := range files {
		if file.Name() == name {
			return true
		}
	}
	return false
}
