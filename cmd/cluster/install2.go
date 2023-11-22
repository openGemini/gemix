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
	"os"
	"path"

	"github.com/openGemini/gemix/pkg/cluster/manager"
	"github.com/openGemini/gemix/pkg/gui"
	"github.com/openGemini/gemix/pkg/utils"
	"github.com/spf13/cobra"
)

func installCmd2() *cobra.Command {
	opt := manager.InstallOptions{
		IdentityFile: path.Join(utils.UserHome(), ".ssh", "id_rsa"),
	}

	cmd := &cobra.Command{
		Use:          "install2 <cluster-name> <version> <topology.yaml>",
		Short:        "Install an openGemini cluster for production",
		Long:         `Install an openGemini cluster for production. SSH connection will be used to deploy files, as well as create system users for running the service.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			shouldContinue, err := gui.CheckCommandArgsAndMayPrintHelp(cmd, args, 3)
			if err != nil {
				return err
			}
			if !shouldContinue {
				return nil
			}

			clusterName := args[0]
			version, err := utils.FmtVer(args[1])
			if err != nil {
				return err
			}

			topoFile := args[2]
			if data, err := os.ReadFile(topoFile); err != nil {
				return err
			} else {
				teleTopology = string(data)
			}

			return cm.Install(clusterName, version, topoFile, opt, skipConfirm, gOpt)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			switch len(args) {
			case 2:
				return nil, cobra.ShellCompDirectiveDefault
			default:
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
		},
	}

	cmd.Flags().StringVarP(&opt.User, "user", "u", "", "The user name to login via SSH. The user must has root (or sudo) privilege.")
	cmd.Flags().BoolVarP(&opt.SkipCreateUser, "skip-create-user", "", false, "Skip creating the user specified in topology.")
	cmd.Flags().StringP("key", "k", "", "The path of the SSH identity file. If specified, public key authentication will be used.")
	cmd.Flags().BoolVarP(&opt.UsePassword, "password", "p", false, "Use password of target hosts. If specified, password authentication will be used.")
	cmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip all confirmations and assumes 'yes'")
	return cmd
}
