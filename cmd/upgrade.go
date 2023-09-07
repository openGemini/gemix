package cmd

import (
	"fmt"
	"openGemini-UP/pkg/deploy"
	"openGemini-UP/pkg/stop"
	"openGemini-UP/util"

	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "upgrade cluster",
	Long:  `upgrade an openGemini cluster to the specified version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("--------------- Cluster upgrading! ---------------")
		version, _ := cmd.Flags().GetString("version")
		if version == "" {
			version = util.Download_version
		}
		fmt.Println("upgrade to cluster version: ", version)

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

		// upload new bin files and start new services
		deployer := deploy.NewGeminiDeployer(version)
		defer deployer.Close()

		if err := deployer.PrepareForDeploy(); err != nil {
			fmt.Println(err)
			return
		}
		if err := deployer.Deploy(); err != nil {
			fmt.Println(err)
		}

		fmt.Println("--------------- Successfully completed cluster upgrade! ---------------")
	},
}

func init() {
	clusterCmd.AddCommand(upgradeCmd)
	upgradeCmd.Flags().StringP("version", "v", "", "component name")
}
