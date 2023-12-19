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
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/openGemini/gemix/pkg/utils"
)

const (
	// Timeout in second when quering node status
	statusQueryTimeout = 10 * time.Second
)

// statusByHost queries current status of the instance by http status api.
func statusByHost(host string, port int, path string, timeout time.Duration, tlsCfg *tls.Config) string {
	if timeout < time.Second {
		timeout = statusQueryTimeout
	}

	client := utils.NewHTTPClient(timeout, tlsCfg)

	scheme := "http"
	if tlsCfg != nil {
		scheme = "https"
	}
	if path == "" {
		path = "/"
	}
	url := fmt.Sprintf("%s://%s%s", scheme, utils.JoinHostPort(host, port), path)

	// body doesn't have any status section needed
	body, err := client.Get(context.TODO(), url)
	if err != nil || body == nil {
		return "Down"
	}
	return "Up"
}

// UptimeByHost queries current uptime of the instance by http Prometheus metric api.
func UptimeByHost(host string, port int, timeout time.Duration, tlsCfg *tls.Config) time.Duration {
	return 0
}
