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

	"github.com/openGemini/gemix/pkg/cluster/operation"
	"github.com/openGemini/gemix/utils"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "install database components",
	Long: `Install database components. The current version supports downloading all database components 
	according to the version number, including ts-meta, ts-sql, ts-store, etc., as well as related configuration files, 
	and save them in the local default path after decompression.`,
	Run: func(cmd *cobra.Command, args []string) {
		version, _ := cmd.Flags().GetString("version")
		if version == "" {
			latestVer, err := utils.GetLatestVerFromCurl()
			if err != nil {
				fmt.Println(err)
				fmt.Println(cmd.UsageString())
				return
			} else {
				version = latestVer
			}
		}
		os, _ := cmd.Flags().GetString("os")
		if os == "" {
			os = utils.DownloadDefaultOs
		}
		arch, _ := cmd.Flags().GetString("arch")
		if arch == "" {
			arch = utils.DownloadDefaultArch
		}
		dOps := operation.DownloadOptions{
			Version: version,
			Os:      os,
			Arch:    arch,
		}

		downloader := operation.NewGeminiDownloader(dOps)
		if err := downloader.Run(); err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(installCmd)
	installCmd.Flags().StringP("version", "v", "", "component version; default is the latest version")
	installCmd.Flags().StringP("os", "o", "", "operating system, supported values: linux/darwin; default is linux")
	installCmd.Flags().StringP("arch", "a", "", "system architecture, supported values: amd64/arm64; default is amd64")
}
