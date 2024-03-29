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

package scripts

import (
	"bytes"
	"path"
	"text/template"

	"github.com/openGemini/gemix/embed"
	"github.com/openGemini/gemix/pkg/utils"
	"github.com/pkg/errors"
)

// TSMonitorScript represent the data to generate ts-monitor config
type TSMonitorScript struct {
	DeployDir string
	LogDir    string
}

// Config generate the config file data.
func (c *TSMonitorScript) Config() ([]byte, error) {
	fp := path.Join("templates", "scripts", "run_ts_monitor.sh.tpl")
	tpl, err := embed.ReadTemplate(fp)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c.ConfigWithTemplate(string(tpl))
}

// ConfigWithTemplate generate the BlackboxExporter config content by tpl
func (c *TSMonitorScript) ConfigWithTemplate(tpl string) ([]byte, error) {
	tmpl, err := template.New("TSMonitorScript").Parse(tpl)
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
func (c *TSMonitorScript) ConfigToFile(file string) error {
	config, err := c.Config()
	if err != nil {
		return errors.WithStack(err)
	}
	return utils.WriteFile(file, config, 0755)
}
