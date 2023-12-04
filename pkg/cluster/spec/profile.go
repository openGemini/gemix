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
	"os/user"
	"path"
	"path/filepath"

	"github.com/openGemini/gemix/pkg/localdata"
	"github.com/pkg/errors"
)

// the subdirectory names
const (
	OpenGeminiPackageCacheDir = "packages"
	OpenGeminiClusterDir      = "clusters"
)

var profileDir string

// getHomeDir gets the home directory of current user (if they have one).
// The result path might be empty.
func getHomeDir() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", errors.WithStack(err)
	}
	return u.HomeDir, nil
}

var initialized = false

// Initialize initializes the global variables of meta package. If the
// environment variable GEMIX_COMPONENT_DATA_DIR is set, it is used as root of
// the profile directory, otherwise the `$HOME/.gemix` of current user is used.
// The directory will be created before return if it does not already exist.
func Initialize(base string) error {
	gemixData := os.Getenv(localdata.EnvNameComponentDataDir)
	gemixHome := os.Getenv(localdata.EnvNameHome)

	switch {
	case gemixData != "":
		profileDir = gemixHome
	case gemixHome != "":
		profileDir = filepath.Join(gemixHome, localdata.StorageParentDir, base)
	default:
		homeDir, err := getHomeDir()
		if err != nil {
			return errors.WithStack(err)
		}
		profileDir = filepath.Join(homeDir, localdata.ProfileDirName, localdata.StorageParentDir, base)
	}

	clusterBaseDir := filepath.Join(profileDir, OpenGeminiClusterDir)
	openGeminiSpec = NewSpec(clusterBaseDir, func() *ClusterMeta {
		return &ClusterMeta{
			Topology: new(Specification),
		}
	})
	initialized = true
	return os.MkdirAll(profileDir, 0750)
}

// ProfileDir returns the full profile directory path of TiUP.
func ProfileDir() string {
	return profileDir
}

// ProfilePath joins a path under the profile dir
func ProfilePath(subpath ...string) string {
	return path.Join(append([]string{profileDir}, subpath...)...)
}

// ClusterPath returns the full path to a subpath (file or directory) of a
// cluster, it is a subdir in the profile dir of the user, with the cluster name
// as its name.
// It is not guaranteed the path already exist.
func ClusterPath(cluster string, subpath ...string) string {
	return GetSpecManager().Path(cluster, subpath...)
}
