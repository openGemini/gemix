/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"openGemini-UP/pkg/stop"

	"github.com/spf13/cobra"
)

// destroyCmd represents the list command
var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "destroy cluster",
	Long:  `destroy an openGemini cluster based on configuration files. Stop all services and delete all logs and data`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("--------------- Cluster destroying! ---------------")
		// stop all services and delete all logs and data
		stop := stop.NewGeminiStop(true)
		defer stop.Close()

		if err := stop.Prepare(); err != nil {
			fmt.Println(err)
			return
		}
		if err := stop.Run(); err != nil {
			fmt.Println(err)
		}
		fmt.Println("--------------- Successfully completed cluster destroy! ---------------")
	},
}

func init() {
	clusterCmd.AddCommand(destroyCmd)
}
