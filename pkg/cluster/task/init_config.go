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

package task

import (
	"context"
	"fmt"
	"os"

	"github.com/openGemini/gemix/pkg/cluster/ctxt"
	"github.com/openGemini/gemix/pkg/cluster/spec"
	"github.com/openGemini/gemix/pkg/meta"
	"github.com/pkg/errors"
)

// InitConfig is used to copy all configurations to the target directory of path
type InitConfig struct {
	specManager    *spec.SpecManager
	clusterName    string
	clusterVersion string
	instance       spec.Instance
	deployUser     string
	ignoreCheck    bool
	paths          meta.DirPaths
}

// Execute implements the Task interface
func (c *InitConfig) Execute(ctx context.Context) error {
	// Copy to remote server
	exec, found := ctxt.GetInner(ctx).GetExecutor(c.instance.GetManageHost())
	if !found {
		return ErrNoExecutor
	}

	if err := os.MkdirAll(c.paths.Cache, 0750); err != nil {
		return errors.WithMessagef(err, "create cache directory failed: %s", c.paths.Cache)
	}

	err := c.instance.InitConfig(ctx, exec, c.clusterName, c.clusterVersion, c.deployUser, c.paths)
	if err != nil {
		return errors.WithMessagef(err, "init config failed: %s:%d", c.instance.GetManageHost(), c.instance.GetPort())
	}
	return nil
}

// Rollback implements the Task interface
func (c *InitConfig) Rollback(ctx context.Context) error {
	return ErrUnsupportedRollback
}

// String implements the fmt.Stringer interface
func (c *InitConfig) String() string {
	return fmt.Sprintf("InitConfig: cluster=%s, user=%s, host=%s, path=%s, %s",
		c.clusterName, c.deployUser, c.instance.GetManageHost(),
		c.specManager.Path(c.clusterName, spec.TempConfigPath, c.instance.ServiceName()),
		c.paths)
}
