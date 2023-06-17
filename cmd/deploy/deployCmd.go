package deploy

import (
	"github.com/spf13/cobra"
	"kubectl-plugin-pod/config"
	"log"
)

var deployCmd = &cobra.Command{
	Use:          "kubectl deploy [flags]",
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("ok")
	},
}

func RunCmd() {
	// 合并主命令的参数
	config.MergeFlags(deployCmd, promptCmd)

	// 加入子命令
	deployCmd.AddCommand(promptCmd)

	if err := deployCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
