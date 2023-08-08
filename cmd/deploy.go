/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"openGemini-UP/pkg/deploy"
	"openGemini-UP/util"

	"github.com/spf13/cobra"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "deploy cluster",
	Long:  `Deploy an openGemini cluster based on configuration files and version numbers.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("--------------- Cluster deploying! ---------------")

		version, _ := cmd.Flags().GetString("version")
		if version == "" {
			version = util.Download_version
		}

		deployer := deploy.NewGeminiDeployer(version)
		defer deployer.Close()

		if err := deployer.PrepareForDeploy(); err != nil {
			fmt.Println(err)
			return
		}
		if err := deployer.Deploy(); err != nil {
			fmt.Println(err)
		}
		fmt.Println("--------------- Successfully completed cluster deployment! ---------------")
	},
}

func init() {
	clusterCmd.AddCommand(deployCmd)
	deployCmd.Flags().StringP("version", "v", "", "component name")
}
