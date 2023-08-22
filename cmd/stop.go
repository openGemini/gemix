/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"openGemini-UP/pkg/stop"

	"github.com/spf13/cobra"
)

// stopCmd represents the list command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop cluster",
	Long:  `Stop an openGemini cluster based on configuration files.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("--------------- Cluster stopping! ---------------")
		// stop all services
		stop := stop.NewGeminiStop(false)
		defer stop.Close()

		if err := stop.Prepare(); err != nil {
			fmt.Println(err)
			return
		}
		if err := stop.Run(); err != nil {
			fmt.Println(err)
		}
		fmt.Println("--------------- Successfully completed cluster stop! ---------------")
	},
}

func init() {
	clusterCmd.AddCommand(stopCmd)
}
