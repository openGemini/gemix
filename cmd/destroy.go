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

		ops, err := getClusterOptions(cmd)
		if err != nil {
			fmt.Println(err)
			return
		}

		// stop all services and delete all logs and data
		stop := stop.NewGeminiStop(ops, true)
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
	destroyCmd.Flags().StringP("yaml", "y", "", "The path to cluster configuration yaml file")
	destroyCmd.Flags().StringP("user", "u", "", "The user name to login via SSH. The user must has root (or sudo) privilege.")
	destroyCmd.Flags().StringP("key", "k", "", "The path of the SSH identity file. If specified, public key authentication will be used.")
	destroyCmd.Flags().StringP("password", "p", "", "The password of target hosts. If specified, password authentication will be used.")
}
