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
	"strings"

	"github.com/fatih/color"
	"github.com/openGemini/gemix/pkg/cluster/manager"
	operator "github.com/openGemini/gemix/pkg/cluster/operation"
	"github.com/openGemini/gemix/pkg/cluster/spec"
	"github.com/openGemini/gemix/pkg/gui"
	"github.com/openGemini/gemix/pkg/logger"
	logprinter "github.com/openGemini/gemix/pkg/logger/printer"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	ClusterCmd   *cobra.Command // represents the cluster command
	gOpt         operator.Options
	teleTopology string //lint:ignore U1000 keep this
	skipConfirm  bool
	log          = logprinter.NewLogger("") // init default logger
)

var openGeminiSpec *spec.SpecManager
var cm *manager.Manager

func init() {
	logger.InitGlobalLogger()

	gui.AddColorFunctionsForCobra()

	cobra.EnableCommandSorting = false

	ClusterCmd = &cobra.Command{
		Use:           "cluster",
		Short:         "Deploy an openGemini cluster for production",
		Long:          `Deploy an openGemini cluster for production`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			log.SetDisplayModeFromString(gOpt.DisplayMode)

			if err := spec.Initialize("cluster"); err != nil {
				return err
			}
			openGeminiSpec = spec.GetSpecManager()
			cm = manager.NewManager("openGemini", openGeminiSpec, log)
			return nil
		},
	}

	ClusterCmd.AddCommand(
		templateCmd(),
		installCmd(),
		//installCmd2(),
		startCmd(),
		//startCmd2,
		stopCmd(),
		//stopCmd2,
		//uninstallCmd,
		newUninstallCmd(),
		statusCmd,
		upgradeCmd,
	)

	//ClusterCmd.PersistentFlags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip all confirmations and assumes 'yes'")
}

// Execute executes the root command
func Execute() {
	zap.L().Info("Execute command", zap.String("command", strings.Join(os.Args, " ")))
	zap.L().Debug("Environment variables", zap.Strings("env", os.Environ()))

	code := 0
	err := ClusterCmd.Execute()
	if err != nil {
		code = 1
	}

	zap.L().Info("Execute command finished", zap.Int("code", code), zap.Error(err))

	if err != nil {
		gui.ColorErrorMsg.Fprintf(os.Stderr, "\nError: %s", err.Error())

		logger.OutputDebugLog("gemix-cluster")
	}

	err = logger.OutputAuditLogIfEnabled()
	if err != nil {
		zap.L().Warn("Write audit log file failed", zap.Error(err))
	}

	color.Unset()

	if code != 0 {
		os.Exit(code)
	}
}
