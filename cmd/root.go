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
	"os"
	"time"

	"github.com/openGemini/gemix/cmd/cluster"
	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "gemix",
	Short: "One-click deployment and upgrade tool for openGemini",
	Long: `Gemix is a command-line component management tool that can help to download and install
	openGemini platform components to the local system. You can run a specific version via
	"gemix install <componet>[:version]". If no version number is specified, the latest version installed
	locally will be used. If the specified component does not have any version installed locally,
	the latest stable version will be downloaded from the repository.`,
}

func Execute() {
	RootCmd.AddCommand(cluster.ClusterCmd)
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var (
	Version   string
	Commit    string
	Branch    string
	BuildTime string
)

func init() {
	// If commit, branch, or build time are not set, make that clear.
	if Version == "" {
		Version = "unknown"
	}
	if Commit == "" {
		Commit = "unknown"
	}
	if Branch == "" {
		Branch = "unknown"
	}

	if BuildTime == "" {
		BuildTime = time.Now().UTC().String()
	}
}
