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
	"path/filepath"

	"github.com/openGemini/gemix/util"
	"github.com/spf13/cobra"
)

var RootCmd *cobra.Command // represents the cluster command

func init() {
	RootCmd = &cobra.Command{
		Use:           "cluster",
		Short:         "Deploy an openGemini cluster for production",
		Long:          `Deploy an openGemini cluster for production`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Example: `
$ gemix cluster install
$ gemix cluster start
$ gemix cluster stop
$ gemix cluster uninstall
`,
		Run: func(cmd *cobra.Command, args []string) {
			//if err = spec.Initialize("cluster"); err != nil {
			//	return err
			//}
			//cm = manager.NewManager("openGemini", openGeminiSpec, spec.openGeminiComponentVersion, log)
		},
	}

	RootCmd.AddCommand(
		installCmd(),
		startCmd,
		stopCmd,
		uninstallCmd,
		statusCmd,
		upgradeCmd,
	)
}

func ReadClusterOptionsByName(cmd *cobra.Command) (util.ClusterOptions, error) {
	var ops util.ClusterOptions
	var err error
	if name, _ := cmd.Flags().GetString("name"); name == "" {
		return ops, fmt.Errorf("the cluster name is required")
	} else if !util.CheckClusterNameExist(name) {
		return ops, fmt.Errorf("the cluster name is not existed, please install the cluster first")
	} else {
		ops, err = util.LoadClusterOptionsFromFile(filepath.Join(util.ClusterInfoDir, name))
		if err != nil {
			return ops, err
		}
	}
	return ops, err
}
