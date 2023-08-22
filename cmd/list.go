/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"io/ioutil"
	"openGemini-UP/util"
	"os"
	"path/filepath"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list of database components",
	Long:  `Display the version information of all components currently downloaded.`,
	Run: func(cmd *cobra.Command, args []string) {
		result := list(util.Download_dst)
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Version", "Component"})

		for key, values := range result {
			row := []string{key, ""}
			for _, value := range values {
				row[1] += value + "\n"
			}
			table.Append(row)
		}

		table.Render()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func list(path string) map[string][]string {
	var result = make(map[string][]string)

	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && isBinPath(path) {
			files, err := ioutil.ReadDir(path)
			if err != nil {
				return err
			}
			version := getBinPathVersion(path)
			for _, file := range files {
				switch file.Name() {
				case "ts-sql":
					result[version] = append(result[version], "ts-sql")
				case "ts-meta":
					result[version] = append(result[version], "ts-meta")
				case "ts-store":
					result[version] = append(result[version], "ts-store")
				case "ts-cli":
					result[version] = append(result[version], "ts-cli")
				case "ts-server":
					result[version] = append(result[version], "ts-server")
				case "ts-monitor":
					result[version] = append(result[version], "ts-monitor")
				}
			}
		}
		return nil
	})
	return result
}

func isBinPath(path string) bool {
	if len(path) > 3 && path[len(path)-3:] == "bin" {
		return true
	}
	return false
}

func getBinPathVersion(path string) string {
	if !isBinPath(path) {
		return ""
	}
	if len(path) <= 14 {
		return ""
	}
	return path[len(path)-14 : len(path)-8]
}
