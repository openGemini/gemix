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

package task

import (
	"context"
	"fmt"

	operator "github.com/openGemini/gemix/pkg/cluster/operation"
	"github.com/pkg/errors"
)

// Downloader is used to download the specific version of a component from
// the repository, there is nothing to do if the specified version exists.
type Downloader struct {
	component string
	os        string
	arch      string
	version   string
}

// NewDownloader create a Downloader instance.
func NewDownloader(component string, os string, arch string, version string) *Downloader {
	return &Downloader{
		component: component,
		os:        os,
		arch:      arch,
		version:   version,
	}
}

// Execute implements the Task interface
func (d *Downloader) Execute(_ context.Context) error {
	// If the version is not specified, the last stable one will be used
	if d.version == "" {
		// TODO: get latest version
		d.version = "latest"
	}

	prefix := fmt.Sprintf("  - Download %s:%s (%s/%s)", d.component, d.version, d.os, d.arch)

	err := operator.Download(prefix, d.component, d.os, d.arch, d.version)
	return errors.WithStack(err)
}

// Rollback implements the Task interface
func (d *Downloader) Rollback(ctx context.Context) error {
	// We cannot delete the component because of some versions maybe exists before
	return nil
}

// String implements the fmt.Stringer interface
func (d *Downloader) String() string {
	return fmt.Sprintf("Download: component=%s, version=%s, os=%s, arch=%s",
		d.component, d.version, d.os, d.arch)
}
