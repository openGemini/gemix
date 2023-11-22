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

	tea "github.com/charmbracelet/bubbletea"
	"github.com/openGemini/gemix/pkg/cluster/spec"
	"github.com/openGemini/gemix/pkg/gui/progress"
	"github.com/openGemini/gemix/utils"
	"github.com/pkg/errors"
)

// Download downloads the specific version of a component from the mirror repository,
// there is nothing to do if the specified version exists.
func Download(prefix, component, nodeOS, arch, version string) (*tea.Program, error) {
	if component == "" {
		return nil, errors.New("component name is not specified")
	}
	if version == "" {
		return nil, errors.Errorf("version is not specified for component '%s'", component)
	}
	if strings.HasPrefix(version, "v") || strings.HasPrefix(version, "V") {
		version = version[1:]
	}

	fileName := fmt.Sprintf("%s-%s-%s-%s.tar.gz", component, version, nodeOS, arch)
	srcPath := spec.ProfilePath(spec.OpenGeminiPackageCacheDir, fileName)
	if err := os.MkdirAll(spec.ProfilePath(spec.OpenGeminiPackageCacheDir), 0750); err != nil {
		return nil, errors.WithStack(err)
	}

	//progress.StartDownload([]string{fileName})

	//lint:ignore SA9003 TODO: verify component sha256
	if utils.IsExist(srcPath) {
		//os.Remove(srcPath)
	}

	// Download from repository if not exists
	if utils.IsNotExist(srcPath) {
		componentUrl := strings.Join([]string{utils.DownloadWeb, "v" + version, fileName}, "/")
		p, err := progress.NewDownloadProgram(prefix, componentUrl, srcPath)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return p, nil
	}
	// component is already downloaded
	return nil, nil
}
