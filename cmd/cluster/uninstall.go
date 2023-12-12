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

package cluster

import (
	operator "github.com/openGemini/gemix/pkg/cluster/operation"
	"github.com/openGemini/gemix/pkg/cluster/spec"
	"github.com/openGemini/gemix/pkg/set"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newUninstallCmd() *cobra.Command {
	destroyOpt := operator.Options{}
	cmd := &cobra.Command{
		Use:   "uninstall <cluster-name>",
		Short: "Uninstall a specified cluster",
		Long: `Uninstall a specified cluster, which will clean the deployment binaries and data.
`,
		//You can retain some nodes and roles data when destroy cluster, see Example:
		//		Example: `
		//$ gemix cluster uninstall <cluster-name> --retain-role-data grafana
		//$ gemix cluster uninstall <cluster-name> --retain-node-data 172.16.13.11:3000
		//$ gemix cluster uninstall <cluster-name> --retain-node-data 172.16.13.12
		//		`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Help()
			}

			clusterName := args[0]

			// Validate the retained roles to prevent unexpected deleting data
			if len(destroyOpt.RetainDataRoles) > 0 {
				roles := set.NewStringSet(spec.AllComponentNames()...)
				for _, role := range destroyOpt.RetainDataRoles {
					if !roles.Exist(role) {
						return errors.Errorf("role name `%s` invalid", role)
					}
				}
			}

			return cm.UninstallCluster(clusterName, gOpt, destroyOpt, skipConfirm)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			switch len(args) {
			case 0:
				return shellCompGetClusterName(cm, toComplete)
			default:
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
		},
	}

	//cmd.Flags().StringArrayVar(&destroyOpt.RetainDataNodes, "retain-node-data", nil, "Specify the nodes or hosts whose data will be retained")
	//cmd.Flags().StringArrayVar(&destroyOpt.RetainDataRoles, "retain-role-data", nil, "Specify the roles whose data will be retained")
	cmd.Flags().BoolVar(&destroyOpt.Force, "force", false, "Force will ignore remote error while destroy the cluster")

	return cmd
}
