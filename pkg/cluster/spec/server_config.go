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
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/openGemini/gemix/pkg/cluster/ctxt"
	"github.com/openGemini/gemix/pkg/meta"
	"github.com/pkg/errors"
)

const (
	// TempConfigPath is the sub path where generated temporary configs are stored
	TempConfigPath = "config-cache"
)

// ErrorCheckConfig represent error occurred in config check stage
var ErrorCheckConfig = errors.New("check config failed")

// strKeyMap tries to convert `map[any]any` to `map[string]any`
func strKeyMap(val any) any {
	m, ok := val.(map[any]any)
	if ok {
		ret := map[string]any{}
		for k, v := range m {
			kk, ok := k.(string)
			if !ok {
				return val
			}
			ret[kk] = strKeyMap(v)
		}
		return ret
	}

	rv := reflect.ValueOf(val)
	if rv.Kind() == reflect.Slice {
		var ret []any
		for i := 0; i < rv.Len(); i++ {
			ret = append(ret, strKeyMap(rv.Index(i).Interface()))
		}
		return ret
	}

	return val
}

func foldKey(key string, val any) (string, any) {
	parts := strings.SplitN(key, ".", 2)
	if len(parts) == 1 {
		return key, strKeyMap(val)
	}
	subKey, subVal := foldKey(parts[1], val)
	return parts[0], map[string]any{
		subKey: strKeyMap(subVal),
	}
}

func patch(origin map[string]any, key string, val any) {
	origVal, found := origin[key]
	if !found {
		origin[key] = strKeyMap(val)
		return
	}
	origMap, lhsOk := origVal.(map[string]any)
	valMap, rhsOk := val.(map[string]any)
	if lhsOk && rhsOk {
		for k, v := range valMap {
			patch(origMap, k, v)
		}
	} else {
		// overwrite
		origin[key] = strKeyMap(val)
	}
}

// FoldMap convert single layer map to multi-layer
func FoldMap(ms map[string]any) map[string]any {
	// we flatten map first to deal with the case like:
	// a.b:
	//   c.d: xxx
	ms = FlattenMap(ms)
	result := map[string]any{}
	for k, v := range ms {
		key, val := foldKey(k, v)
		patch(result, key, val)
	}
	return result
}

// FlattenMap convert mutil-layer map to single layer
func FlattenMap(ms map[string]any) map[string]any {
	result := map[string]any{}
	for k, v := range ms {
		var sub map[string]any

		if m, ok := v.(map[string]any); ok {
			sub = FlattenMap(m)
		} else if m, ok := v.(map[any]any); ok {
			fixM := map[string]any{}
			for k, v := range m {
				if sk, ok := k.(string); ok {
					fixM[sk] = v
				}
			}
			sub = FlattenMap(fixM)
		} else {
			result[k] = v
			continue
		}

		for sk, sv := range sub {
			result[k+"."+sk] = sv
		}
	}
	return result
}

// MergeConfig merge two or more config into one and unflat them
// config1:
//
//	a.b.a: 1
//	a.b.b: 2
//
// config2:
//
//	a.b.a: 3
//	a.b.c: 4
//
// config3:
//
//	b.c = 5
//
// After MergeConfig(config1, config2, config3):
//
//	a:
//	  b:
//	    a: 3
//	    b: 2
//	    c: 4
//	b:
//	  c: 5
func MergeConfig(orig map[string]any, overwrites ...map[string]any) map[string]any {
	lhs := FoldMap(orig)
	for _, overwrite := range overwrites {
		rhs := FoldMap(overwrite)
		for k, v := range rhs {
			patch(lhs, k, v)
		}
	}
	return lhs
}

// GetValueFromPath try to find the value by path recursively
func GetValueFromPath(m map[string]any, p string) any {
	ss := strings.Split(p, ".")

	searchMap := make(map[any]any)
	m = FoldMap(m)
	for k, v := range m {
		searchMap[k] = v
	}

	return searchValue(searchMap, ss)
}

func searchValue(m map[any]any, ss []string) any {
	l := len(ss)
	switch l {
	case 0:
		return m
	case 1:
		return m[ss[0]]
	}

	key := ss[0]
	if pm, ok := m[key].(map[any]any); ok {
		return searchValue(pm, ss[1:])
	} else if pm, ok := m[key].(map[string]any); ok {
		searchMap := make(map[any]any)
		for k, v := range pm {
			searchMap[k] = v
		}
		return searchValue(searchMap, ss[1:])
	}

	return nil
}

// Merge2Toml merge the config of global.
func Merge2Toml(comp string, global, overwrite map[string]any) ([]byte, error) {
	lhs := MergeConfig(global, overwrite)
	buf := bytes.NewBufferString(fmt.Sprintf(`# WARNING: This file is auto-generated. Do not edit! All your modification will be overwritten!
# You can use 'gemix cluster edit-config' and 'gemix cluster reload' to update the configuration
# All configuration items you want to change can be added to:
# server_configs:
#   %s:
#     aa.b1.c3: value
#     aa.b2.c4: value
`, comp))

	enc := toml.NewEncoder(buf)
	enc.Indent = ""
	err := enc.Encode(lhs)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return buf.Bytes(), nil
}

// BindVersion map the cluster version to the third components binding version.
type BindVersion func(comp string, version string) (bindVersion string)

func checkConfig(ctx context.Context, e ctxt.Executor, componentName, componentSource, clusterVersion, nodeOS, arch, config string, paths meta.DirPaths, bindVersion BindVersion) error {
	// TODO:
	return nil
}
