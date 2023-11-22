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
	"fmt"

	"github.com/openGemini/gemix/utils"
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "upgrade cluster",
	Long:  `upgrade an openGemini cluster to the specified version`,
	Run: func(cmd *cobra.Command, args []string) {
		var ops utils.ClusterOptions
		var err error
		if ops, err = ReadClusterOptionsByName(cmd); err != nil {
			fmt.Println(err)
			fmt.Println(cmd.UsageString())
			return
		}

		var new_version string
		if new_version, _ = cmd.Flags().GetString("version"); new_version == "" {
			fmt.Println("the new_version is required")
			fmt.Println(cmd.UsageString())
			return
		}

		err = UpgradeCluster(ops, new_version)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func UpgradeCluster(ops utils.ClusterOptions, newV string) error {
	newOps := ops
	newOps.Version = newV

	// stop all services
	if err := StopCluster(ops); err != nil {
		return err
	}

	// uninstall openGeini
	if err := UninstallCluster(ops); err != nil {
		return err
	}

	// install new cluster
	if err := InstallCluster(newOps); err != nil {
		return err
	}

	// start new cluster
	if err := StartCluster(newOps); err != nil {
		return err
	}
	fmt.Printf("Successfully upgraded the openGemini cluster from %s to %s\n", ops.Version, newV)
	return nil
}

func init() {
	upgradeCmd.Flags().StringP("name", "n", "", "cluster name")
	upgradeCmd.Flags().StringP("version", "v", "", "new cluster version")
}
