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

// statusCmd
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "check cluster status",
	Long:  `Check the current running status of an openGemini cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		var ops util.ClusterOptions
		var err error
		if ops, err = ReadClusterOptionsByName(cmd); err != nil {
			fmt.Println(err)
			fmt.Println(cmd.UsageString())
			return
		}

		err = PatrolCluster(ops)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func PatrolCluster(ops util.ClusterOptions) error {
	patroler := manager.NewGeminiStatusPatroller(ops)
	defer patroler.Close()

	if err := patroler.PrepareForPatrol(); err != nil {
		return err
	}
	if err := patroler.Patrol(); err != nil {
		return err
	}
	return nil
}

func init() {
	statusCmd.Flags().StringP("name", "n", "", "cluster name")
}
