package cmd

import (
	"github.com/spf13/cobra"
	"kubectl-plugin-pod/config"
	"log"
)

var rootCmd = &cobra.Command{
	Use:          "kubectl pods [flags]",
	SilenceUsage: true,
}

func RunCmd() {
	config.Clientset = config.NewK8sConfig().InitClient()

	// 合并主命令的参数
	config.MergeFlags(rootCmd, podListCmd, promptCmd)

	// 加载pod列表flag
	addListFlags()

	// 加入子命令
	rootCmd.AddCommand(podListCmd, promptCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
