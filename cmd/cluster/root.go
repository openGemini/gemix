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

	"github.com/openGemini/gemix/pkg/cluster/manager"
	operator "github.com/openGemini/gemix/pkg/cluster/operation"
	"github.com/openGemini/gemix/pkg/cluster/spec"
	"github.com/openGemini/gemix/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	RootCmd      *cobra.Command // represents the cluster command
	gOpt         operator.Options
	teleTopology string //lint:ignore U1000 keep this
	skipConfirm  bool
	log          = zap.NewNop() // init default logger
)

var openGeminiSpec *spec.SpecManager
var cm *manager.Manager

func init() {
	RootCmd = &cobra.Command{
		Use:           "cluster",
		Short:         "Deploy an openGemini cluster for production",
		Long:          `Deploy an openGemini cluster for production`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := spec.Initialize("cluster"); err != nil {
				return err
			}
			openGeminiSpec = spec.GetSpecManager()
			cm = manager.NewManager("openGemini", openGeminiSpec, log)
			return nil
		},
	}

	RootCmd.AddCommand(
		installCmd(),
		installCmd2(),
		startCmd,
		stopCmd,
		uninstallCmd,
		statusCmd,
		upgradeCmd,
	)

	//RootCmd.PersistentFlags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip all confirmations and assumes 'yes'")
}

func ReadClusterOptionsByName(cmd *cobra.Command) (utils.ClusterOptions, error) {
	var ops utils.ClusterOptions
	var err error
	if name, _ := cmd.Flags().GetString("name"); name == "" {
		return ops, fmt.Errorf("the cluster name is required")
	} else if !utils.CheckClusterNameExist(name) {
		return ops, fmt.Errorf("the cluster name is not existed, please install the cluster first")
	} else {
		ops, err = utils.LoadClusterOptionsFromFile(filepath.Join(utils.ClusterInfoDir, name))
		if err != nil {
			return ops, err
		}
	}
	return ops, err
}
