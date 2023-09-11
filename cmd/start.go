/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"openGemini-UP/pkg/deploy"

	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start cluster",
	Long:  `Start an openGemini cluster based on configuration files and version numbers after this cluster was stopped.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("--------------- Cluster starting! ---------------")

		deployer := deploy.NewGeminiDeployer("")
		defer deployer.Close()

		if err := deployer.PrepareForStart(); err != nil {
			fmt.Println(err)
			return
		}
		if err := deployer.Start(); err != nil {
			fmt.Println(err)
		}
		fmt.Println("--------------- Successfully completed cluster start! ---------------")
	},
}

func init() {
	clusterCmd.AddCommand(startCmd)
}
