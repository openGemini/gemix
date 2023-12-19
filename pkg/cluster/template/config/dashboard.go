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

package config

import (
	"bytes"
	"io/fs"
	"path/filepath"
	"text/template"

	"github.com/openGemini/gemix/embed"
	"github.com/openGemini/gemix/pkg/utils"
	"github.com/pkg/errors"
)

// DashboardConfig represent the data to generate Dashboard config
type DashboardConfig struct {
	ClusterName string
	DeployDir   string
}

// NewDashboardConfig returns a DashboardConfig
func NewDashboardConfig(cluster, deployDir string) *DashboardConfig {
	return &DashboardConfig{
		ClusterName: cluster,
		DeployDir:   deployDir,
	}
}

// Config generate the config file data.
func (c *DashboardConfig) Config() ([]byte, error) {
	fp := filepath.Join("templates", "config", "dashboard.yml.tpl")
	tpl, err := embed.ReadTemplate(fp)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c.ConfigWithTemplate(string(tpl))
}

// ConfigWithTemplate generate the Dashboard config content by tpl
func (c *DashboardConfig) ConfigWithTemplate(tpl string) ([]byte, error) {
	tmpl, err := template.New("dashboard").Parse(tpl)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	content := bytes.NewBufferString("")
	if err = tmpl.Execute(content, c); err != nil {
		return nil, errors.WithStack(err)
	}

	return content.Bytes(), nil
}

// ConfigToFile write config content to specific path
func (c *DashboardConfig) ConfigToFile(file string) error {
	config, err := c.Config()
	if err != nil {
		return errors.WithStack(err)
	}
	return utils.WriteFile(file, config, 0750)
}

// ReadDashboardsDir reads config content to specific path
func (c *DashboardConfig) ReadDashboardsDir() ([]fs.DirEntry, error) {
	name := filepath.Join("templates", "dashboards")
	return embed.ReadTemplateDir(name)
}

// DashboardToFile writes config content to specific path
func (c *DashboardConfig) DashboardToFile(file string, tplFile string) error {
	tpl, err := embed.ReadTemplate(tplFile)
	if err != nil {
		return errors.WithStack(err)
	}
	content, err := c.ConfigWithTemplate(string(tpl))
	if err != nil {
		return errors.WithStack(err)
	}

	return utils.WriteFile(file, content, 0750)
}
