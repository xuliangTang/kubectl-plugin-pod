package pod

import (
	"github.com/spf13/cobra"
	"kubectl-plugin-pod/config"
	"log"
)

var podCmd = &cobra.Command{
	Use:          "kubectl pods [flags]",
	SilenceUsage: true,
}

func RunCmd() {
	// 合并主命令的参数
	config.MergeFlags(podCmd, podListCmd, promptCmd)

	// 加载pod列表flag
	addListFlags()

	// 加入子命令
	podCmd.AddCommand(podListCmd, promptCmd)

	if err := podCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
