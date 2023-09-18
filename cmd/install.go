/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"openGemini-UP/pkg/download"
	"openGemini-UP/util"

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
			version = util.Download_default_version
		}
		os, _ := cmd.Flags().GetString("os")
		if os == "" {
			os = util.Download_default_os
		}
		arch, _ := cmd.Flags().GetString("arch")
		if arch == "" {
			arch = util.Download_default_arch
		}
		dOps := download.DownloadOptions{
			Version: version,
			Os:      os,
			Arch:    arch,
		}

		downloader := download.NewGeminiDownloader(dOps)
		if err := downloader.Run(); err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().StringP("version", "v", "", "component version; default is v1.0.0")
	installCmd.Flags().StringP("os", "o", "", "operating system, linux/darwin/windows; default is linux")
	installCmd.Flags().StringP("arch", "a", "", "Supported values: amd64, arm64; default is amd64")
}
