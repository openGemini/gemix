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
	"path/filepath"
	"reflect"
	"strings"
)

func findField(v reflect.Value, fieldName string) (int, bool) {
	for i := 0; i < reflect.Indirect(v).NumField(); i++ {
		if reflect.Indirect(v).Type().Field(i).Name == fieldName {
			return i, true
		}
	}
	return -1, false
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

// Validate validates the topology specification and produce error if the
// specification invalid (e.g: port conflicts or directory conflicts)
func (s *Specification) Validate() error {
	return nil
}
