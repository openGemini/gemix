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

package operation

import (
	"fmt"
	"os"
	"strings"

	"github.com/openGemini/gemix/pkg/cluster/spec"
	ver "github.com/openGemini/gemix/pkg/cluster/version"
	"github.com/openGemini/gemix/pkg/gui/progress"
	"github.com/openGemini/gemix/pkg/repository"
	utils2 "github.com/openGemini/gemix/pkg/utils"
	"github.com/pkg/errors"
)

// Download downloads the specific version of a component from the mirror repository,
// there is nothing to do if the specified version exists.
func Download(prefix, component, nodeOS, arch, version string) error {
	if component == "" {
		return errors.New("component name is not specified")
	}
	if version == "" {
		return errors.Errorf("version is not specified for component '%s'", component)
	}
	if strings.HasPrefix(version, "v") || strings.HasPrefix(version, "V") {
		version = version[1:]
	}

	fileName := fmt.Sprintf("%s-%s-%s-%s.tar.gz", component, version, nodeOS, arch)
	componentUrl := strings.Join([]string{repository.GetRepo(), "v" + version, fileName}, "/")

	if component == spec.ComponentGrafana {
		// FIXME: download from opengemini.org
		if nodeOS == "darwin" {
			arch = "amd64"
		}
		fileName = fmt.Sprintf("%s-enterprise-%s.%s-%s.tar.gz", component, ver.GrafanaVersion, nodeOS, arch)
		componentUrl = strings.Join([]string{"https://dl.grafana.com/oss/release", fileName}, "/")
	}
	dstPath := spec.ProfilePath(spec.OpenGeminiPackageCacheDir, fileName)
	if err := os.MkdirAll(spec.ProfilePath(spec.OpenGeminiPackageCacheDir), 0750); err != nil {
		return errors.WithStack(err)
	}

	if utils2.IsExist(dstPath) {
		if component == spec.ComponentOpenGemini {
			if err := repository.VerifyComponent(version, dstPath); err != nil {
				// TODO: do not remote package
				//_ = os.Remove(dstPath) // nolint
			}
		}
	}

	// Download from repository if not exists
	if utils2.IsNotExist(dstPath) {
		err := progress.NewDownloadProgram(prefix, componentUrl, dstPath)
		if err != nil {
			return errors.WithStack(err)
		}
		return nil
	}
	// component is already downloaded
	return nil
}
