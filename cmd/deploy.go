/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"openGemini-UP/pkg/deploy"

	"github.com/spf13/cobra"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "deploy cluster",
	Long:  `Deploy an openGemini cluster based on configuration files and version numbers.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("--------------- Cluster deploying! ---------------")

		ops, err := getClusterOptions(cmd)
		if err != nil {
			fmt.Println(err)
			return
		}

		deployer := deploy.NewGeminiDeployer(ops)
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
	deployCmd.Flags().StringP("yaml", "y", "", "The path to cluster configuration yaml file")
	deployCmd.Flags().StringP("user", "u", "", "The user name to login via SSH. The user must has root (or sudo) privilege.")
	deployCmd.Flags().StringP("key", "k", "", "The path of the SSH identity file. If specified, public key authentication will be used.")
	deployCmd.Flags().StringP("password", "p", "", "The password of target hosts. If specified, password authentication will be used.")
}
