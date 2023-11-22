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

package executor

import (
	"time"

	"github.com/joomcode/errorx"
)

var (
	errNS = errorx.NewNamespace("executor")

	executeDefaultTimeout = time.Minute
)

// New create a new Executor
func New(sudo bool, c SSHConfig) (*EasySSHExecutor, error) {
	// set default values
	if c.Port <= 0 {
		c.Port = 22
	}

	if c.Timeout == 0 {
		c.Timeout = time.Second * 5 // default timeout is 5 sec
	}

	e := &EasySSHExecutor{
		Locale: "C",
		Sudo:   sudo,
	}
	e.initialize(c)
	return e, nil
}
