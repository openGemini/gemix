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
	"github.com/openGemini/gemix/utils"
	"github.com/spf13/cobra"
)

//lint:ignore U1000 keep this
var startCmd2 = &cobra.Command{
	Use:   "start2",
	Short: "Start an openGemini cluster",
	Long:  `Start an openGemini cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		var ops utils.ClusterOptions
		var err error
		if ops, err = ReadClusterOptionsByName(cmd); err != nil {
			fmt.Println(err)
			fmt.Println(cmd.UsageString())
			return
		}

		err = StartCluster(ops)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("\nCheck the status of openGemini cluster\n")
		err = PatrolCluster(ops)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func StartCluster(clusterOpts utils.ClusterOptions) error {
	starter := manager.NewGeminiStarter(clusterOpts, startOpts)
	defer starter.Close()

	if err := starter.PrepareForStart(); err != nil {
		return err
	}
	if err := starter.Start(); err != nil {
		return err
	}
	fmt.Printf("Successfully started the openGemini cluster with version : %s\n", clusterOpts.Version)
	return nil
}

func init() {
	startCmd2.Flags().StringP("name", "n", "", "cluster name")
	startCmd2.Flags().BoolVarP(&startOpts.SkipCreateUser, "skip-create-user", "", false, "(EXPERIMENTAL) Skip creating the user specified in topology.")
}
