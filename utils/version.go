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

package utils

import (
	"fmt"
	"net/url"
	"os/exec"
	"path"
	"regexp"
	"strings"
)

func GetLatestVerFromCurl() (string, error) {
	cmd := exec.Command("curl", "-i", "-k", DownloadLatestUrl)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	response := string(output)

	re := regexp.MustCompile(`location:\s*(.*?)\n`)
	matches := re.FindStringSubmatch(response)

	if len(matches) < 2 {
		return "", fmt.Errorf("location header not found")
	}
	locationValue := strings.TrimSpace(matches[1])

	parsedURL, err := url.Parse(locationValue)
	if err != nil {
		return "", err
	}
	urlPath := parsedURL.Path
	lastSegment := path.Base(urlPath)

	return lastSegment, nil
}
