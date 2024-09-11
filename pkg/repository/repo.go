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

package repository

import (
	"bufio"
	"os"
	"strings"

	"github.com/openGemini/gemix/pkg/cluster/spec"
	"github.com/openGemini/gemix/pkg/gui/progress"
	"github.com/openGemini/gemix/pkg/utils"
	"github.com/pkg/errors"
)

const (
	GITEE_REPO  = "https://gitee.com/opengemini/Releases/releases/download"    // https://gitee.com/opengemini/Releases/releases/download/v1.1.1/openGemini-1.1.1-linux-amd64.tar.gz
	GITHUB_REPO = "https://github.com/openGemini/openGemini/releases/download" // https://github.com/openGemini/openGemini/releases/download/v1.1.1/openGemini-1.1.1-linux-amd64.tar.gz
	CHECKSUMS   = "checksums.txt"
)

func GetRepo() string {
	repoUrl := os.Getenv("GEMIX_MIRRORS_REPO")
	if strings.Contains(repoUrl, "gitee") {
		return GITEE_REPO
	}
	return GITHUB_REPO
}

func tryToDownloadCheckSumsFile(version string) error {
	dstPath := spec.ProfilePath(spec.OpenGeminiPackageCacheDir, CHECKSUMS)
	checksumsFile := strings.Join([]string{GetRepo(), "v" + version, CHECKSUMS}, "/")
	err := progress.NewDownloadProgram("", checksumsFile, dstPath)
	return errors.WithStack(err)
}

func verifySum256(target string, sha string) error {
	file, err := os.Open(target)
	if err != nil {
		return errors.WithStack(err)
	}
	defer file.Close()

	err = utils.CheckSHA256(file, sha)
	return errors.WithStack(err)
}

func VerifyComponent(version, target string) error {
	checksums := spec.ProfilePath(spec.OpenGeminiPackageCacheDir, CHECKSUMS)
	if utils.IsNotExist(checksums) {
		if err := tryToDownloadCheckSumsFile(version); err != nil {
			// FIXME: assume that the user downloaded the component without problems
			return nil
		}
	}

	file, err := os.Open(checksums)
	if err != nil {
		return err
	}
	defer file.Close()

	var verify bool
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		data := strings.Split(scanner.Text(), " ") // [8b3c2213c044f07108557f899d83ba471ea63455d17113d96c3216e5e2b5fffa, "", openGemini-1.1.1-darwin-amd64.tar.gz]
		if len(data) != 3 {
			_ = os.Remove(checksums) //nolint
			break
		}
		if strings.Contains(target, data[2]) {
			if err = verifySum256(target, data[0]); err != nil {
				return errors.WithStack(err)
			}
			verify = true
			break
		}
	}

	if verify {
		return nil
	}
	return errors.Errorf("failed to verify component")
}
