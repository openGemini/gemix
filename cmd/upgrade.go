package cmd

import (
	"fmt"
	"openGemini-UP/pkg/deploy"
	"openGemini-UP/pkg/stop"

	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "upgrade cluster",
	Long:  `upgrade an openGemini cluster to the specified version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("--------------- Cluster upgrading! ---------------")

		ops, err := getClusterOptions(cmd)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("upgrade to cluster version: ", ops.Version)

		// stop all services
		stop := stop.NewGeminiStop(ops, false)
		defer stop.Close()
		if err := stop.Prepare(); err != nil {
			fmt.Println(err)
			return
		}
		if err := stop.Run(); err != nil {
			fmt.Println(err)
		}

		// upload new bin files and start new services
		deployer := deploy.NewGeminiDeployer(ops)
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
	upgradeCmd.Flags().StringP("yaml", "y", "", "The path to cluster configuration yaml file")
	upgradeCmd.Flags().StringP("user", "u", "", "The user name to login via SSH. The user must has root (or sudo) privilege.")
	upgradeCmd.Flags().StringP("key", "k", "", "The path of the SSH identity file. If specified, public key authentication will be used.")
	upgradeCmd.Flags().StringP("password", "p", "", "The password of target hosts. If specified, password authentication will be used.")
}
