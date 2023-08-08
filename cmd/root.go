/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "openGemini-UP",
	Short: "One-click deployment and upgrade tool for openGemini",
	Long: `openGemini-UP is a one-click deployment and upgrade openGemini tool for users. 
	It can help users easily deploy openGemini clusters based on configuration files written by users.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
