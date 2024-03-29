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
	"net"
	"strconv"
	"strings"
)

// JoinHostPort return host and port
func JoinHostPort(host string, port int) string {
	return net.JoinHostPort(host, strconv.Itoa(port))
}

// Ternary operator
func Ternary(condition bool, a, b any) any {
	if condition {
		return a
	}
	return b
}

// ParseHostPort Prase host and port
func ParseHostPort(hostport string) (host, port string) {
	colon := strings.LastIndex(hostport, ":")

	host = strings.TrimSuffix(strings.TrimPrefix(hostport[:colon], "["), "]")
	port = hostport[colon+1:]
	return
}
