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

	"github.com/openGemini/gemix/pkg/cluster/spec"
	"github.com/pkg/errors"
)

// CopyComponent is used to copy all files related the specific version a component
// to the target directory of path
type CopyComponent struct {
	component  string // component name like "ts-meta/ts-sql/ts-store/..."
	srcPkgName string // src package name like "openGemini"
	os         string
	arch       string
	version    string
	host       string
	srcPath    string
	dstDir     string
}

// Execute implements the Task interface
func (c *CopyComponent) Execute(ctx context.Context) error {
	if c.version == "" {
		// TODO: If the version is not specified, the last stable one will be used
		return errors.New("please specify version")
	} else {
		// rename v1.1.1 to 1.1.1
		c.version = c.version[1:]
	}

	// Copy to remote server
	srcPath := c.srcPath
	if srcPath == "" {
		srcPath = spec.PackagePath(c.srcPkgName, c.version, c.os, c.arch)
	}

	install := &InstallPackage{
		component: c.component,
		srcPath:   srcPath,
		host:      c.host,
		dstDir:    c.dstDir,
	}

	return install.Execute(ctx)
}

// Rollback implements the Task interface
func (c *CopyComponent) Rollback(ctx context.Context) error {
	return ErrUnsupportedRollback
}

// String implements the fmt.Stringer interface
func (c *CopyComponent) String() string {
	return fmt.Sprintf("CopyComponent: component=%s, version=%s, remote=%s:%s os=%s, arch=%s",
		c.component, c.version, c.host, c.dstDir, c.os, c.arch)
}
