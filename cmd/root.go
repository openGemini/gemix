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

package cmd

import (
	"github.com/openGemini/gemix/cmd/cluster"
	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "gemix",
	Short: "One-click deployment and upgrade tool for openGemini",
	Long: `Gemix is a command-line component management tool that can help to download and install
	openGemini platform components to the local system. You can run a specific version via
	"gemix install <component>[:version]". If no version number is specified, the latest version installed
	locally will be used. If the specified component does not have any version installed locally,
	the latest stable version will be downloaded from the repository.`,
	SilenceErrors:      true,
	SilenceUsage:       true,
	DisableFlagParsing: true,
	Args: func(cmd *cobra.Command, args []string) error {
		// Support `gemix <component>`
		return nil
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		switch args[0] {
		case "--help", "-h":
			return cmd.Help()
		case "--version", "-v":
			versionCmd.Run(cmd, nil)
			return nil
		default:
			return cmd.Help()
		}
	},
}

func Execute() {
	RootCmd.AddCommand(cluster.ClusterCmd)
	cluster.Execute()
}
