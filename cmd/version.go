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
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version   string
	Commit    string
	Branch    string
	BuildTime string
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of gemix",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("version: %s, branch: %s, commit: %s, build time: %s\n", Version, Branch, Commit, BuildTime)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
	RootCmd.Flags().BoolP("version", "v", false, "Print the version of gemix")
}
