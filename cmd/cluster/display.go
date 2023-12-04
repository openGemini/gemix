package cluster

import (
	"github.com/openGemini/gemix/pkg/cluster/manager"
	"github.com/spf13/cobra"
)

func shellCompGetClusterName(cm *manager.Manager, toComplete string) ([]string, cobra.ShellCompDirective) {
	var result []string
	//clusters, _ := cm.GetClusterList()
	//for _, c := range clusters {
	//	if strings.HasPrefix(c.Name, toComplete) {
	//		result = append(result, c.Name)
	//	}
	//}
	return result, cobra.ShellCompDirectiveNoFileComp
}
