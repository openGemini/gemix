// Copyright 2023 Huawei Cloud Computing Technologies Co., Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list of available components",
	Long:  `Display the available components of Gemix.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available components:")
		table := tablewriter.NewWriter(os.Stdout)
		data := [][]string{
			{"cluster", "openGemini", "Deploy an openGemini cluster for production"},
			{"opengemini-dashboard", "openGemini", "TODO"},
		}
		table.SetColWidth(100)
		table.SetHeader([]string{"Name", "Owner", "Description"})
		for _, row := range data {
			table.Append(row)
		}
		table.Render()
	},
}

func init() {
	RootCmd.AddCommand(listCmd)
}
