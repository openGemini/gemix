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

	"github.com/openGemini/gemix/pkg/cluster/manager"
	"github.com/openGemini/gemix/util"
	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop cluster",
	Long:  `Stop an openGemini cluster based on configuration files.`,
	Run: func(cmd *cobra.Command, args []string) {
		var ops util.ClusterOptions
		var err error
		if ops, err = ReadClusterOptionsByName(cmd); err != nil {
			fmt.Println(err)
			fmt.Println(cmd.UsageString())
			return
		}

		err = StopCluster(ops)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func StopCluster(ops util.ClusterOptions) error {
	stop := manager.NewGeminiStop(ops)
	defer stop.Close()

	if err := stop.Prepare(); err != nil {
		return err
	}
	if err := stop.Run(); err != nil {
		return err
	}
	fmt.Printf("Successfully stopped the openGemini cluster\n")
	return nil
}

func init() {
	ClusterCmd.AddCommand(stopCmd)
	stopCmd.Flags().StringP("name", "n", "", "cluster name")
}
