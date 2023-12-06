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

package cluster

import (
	"bytes"
	"fmt"
	"path"
	"text/template"

	"github.com/openGemini/gemix/embed"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// TemplateOptions contains the options for print topology template.
type TemplateOptions struct {
	Full  bool // print full template
	Local bool // print and render local template
}

// LocalTemplate contains the variables for print local template.
type LocalTemplate struct {
	GlobalUser      string   // global.user in yaml template
	GlobalGroup     string   // global.group in yaml template
	GlobalSSHPort   int      // global.ssh_port in yaml template
	GlobalDeployDir string   // global.deploy_dir in yaml template
	GlobalLogDir    string   // global.deploy_dir in yaml template
	GlobalDataDir   string   // global.data_dir in yaml template
	GlobalArch      string   // global.arch in yaml template
	TSMetaServers   []string // ts-meta-servers in yaml template
	TSSqlServers    []string // ts-sql-servers in yaml template
	TSStoreServers  []string // ts-store-servers in yaml template
	GrafanaServers  []string // grafana_servers in yaml template
}

// This is used to identify how many bool type options are set, so that an
// error can be throw if more than one is given.
func sumBool(b ...bool) int {
	n := 0
	for _, v := range b {
		if v {
			n++
		}
	}
	return n
}

func templateCmd() *cobra.Command {
	opt := TemplateOptions{}
	localOpt := LocalTemplate{}

	cmd := &cobra.Command{
		Use:   "template",
		Short: "Print topology template",
		RunE: func(cmd *cobra.Command, args []string) error {
			if sumBool(opt.Full, opt.Local) > 1 {
				return errors.New("at most one of 'full', 'multi-dc', or 'local' can be specified")
			}
			name := "minimal.yaml"
			switch {
			case opt.Full:
				name = "topology.example.yaml"
			case opt.Local:
				name = "local.tpl"
			}

			fp := path.Join("examples", "cluster", name)
			tpl, err := embed.ReadExample(fp)
			if err != nil {
				return err
			}

			if !opt.Local {
				// print example yaml and return
				fmt.Fprintln(cmd.OutOrStdout(), string(tpl))
				return nil
			}

			// redner template

			// validate arch
			if localOpt.GlobalArch != "amd64" && localOpt.GlobalArch != "arm64" {
				return fmt.Errorf(`supported values are "amd64" or "arm64" in global.arch`)
			}

			tmpl, err := template.New(name).Parse(string(tpl))
			if err != nil {
				return err
			}

			content := bytes.NewBufferString("")
			if err := tmpl.Execute(content, &localOpt); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), content.String())
			return nil
		},
	}

	cmd.Flags().BoolVar(&opt.Full, "full", false, "Print the full topology template for openGemini cluster.")
	cmd.Flags().BoolVar(&opt.Local, "local", false, "Print and render template for deploying a simple cluster locally.")

	// template values for rendering
	cmd.Flags().StringVar(&localOpt.GlobalUser, "user", "gemini", "The user who runs the openGemini cluster.")
	cmd.Flags().StringVar(&localOpt.GlobalGroup, "group", "", "group is used to specify the group name the user belong to if it's not the same as user.")
	cmd.Flags().IntVar(&localOpt.GlobalSSHPort, "ssh-port", 22, "SSH port of servers in the managed cluster.")
	cmd.Flags().StringVar(&localOpt.GlobalDeployDir, "deploy-dir", "/var/lib/openGemini/deploy", "Storage directory for cluster deployment files, startup scripts, and configuration files.")
	cmd.Flags().StringVar(&localOpt.GlobalLogDir, "log-dir", "/var/lib/openGemini/logs", "Log directory for cluster components.")
	cmd.Flags().StringVar(&localOpt.GlobalDataDir, "data-dir", "/var/lib/openGemini/data", "openGemini Cluster data storage directory.")
	cmd.Flags().StringVar(&localOpt.GlobalArch, "arch", "amd64", "Supported values: \"amd64\", \"arm64\".")
	cmd.Flags().StringSliceVar(&localOpt.TSMetaServers, "ts-meta-servers", []string{"127.0.0.1"}, "List of ts-meta servers")
	cmd.Flags().StringSliceVar(&localOpt.TSSqlServers, "ts-sql-servers", []string{"127.0.0.1"}, "List of ts-sql servers")
	cmd.Flags().StringSliceVar(&localOpt.TSStoreServers, "ts-store-servers", []string{"127.0.0.1"}, "List of ts-store servers")
	cmd.Flags().StringSliceVar(&localOpt.GrafanaServers, "grafana-servers", []string{"127.0.0.1"}, "List of grafana servers")
	return cmd
}
