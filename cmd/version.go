/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "openGemini-UP version",
	Long: `Display the version number of the management deployment tool openGemini-UP, 
	currently there is only one version, the default is v0.0.1.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("openGemini-UP version 0.01")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
