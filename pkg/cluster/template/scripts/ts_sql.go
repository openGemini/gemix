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
	"os"
	"path"
	"text/template"

	"github.com/openGemini/gemix/embed"
)

// TSSqlScript represent the data to generate ts-sql config
type TSSqlScript struct {
	DeployDir string
	LogDir    string
}

// ConfigToFile write config content to specific path
func (c *TSSqlScript) ConfigToFile(file string) error {
	fp := path.Join("templates", "scripts", "run_ts_sql.sh.tpl")
	tpl, err := embed.ReadTemplate(fp)
	if err != nil {
		return err
	}

	tmpl, err := template.New("TSSql").Parse(string(tpl))
	if err != nil {
		return err
	}

	content := bytes.NewBufferString("")
	if err := tmpl.Execute(content, c); err != nil {
		return err
	}

	return os.WriteFile(file, content.Bytes(), 0750)
}
