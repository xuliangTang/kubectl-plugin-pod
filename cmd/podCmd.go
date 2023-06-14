package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"kubectl-plugin-pod/config"
	"log"
)

var clientset *kubernetes.Clientset

var rootCmd = &cobra.Command{
	Use:          "kubectl pods [flags]",
	SilenceUsage: true,
}

func RunCmd() {
	clientset = config.NewK8sConfig().InitClient()

	// 合并主命令的参数
	config.MergeFlags(rootCmd, podListCmd)
	// 加入子命令
	rootCmd.AddCommand(podListCmd)

	podListCmd.Flags().BoolVar(&showLabels, "show-labels", false, "kubectl pods --show-labels")
	podListCmd.Flags().StringVar(&labels, "labels", "", "kubectl pods --labels=\"app=test,version=v1\"")
	podListCmd.Flags().StringVar(&fields, "fields", "", "kubectl pods --fields=\"status.phase=Running\"")
	podListCmd.Flags().StringVar(&name, "name", "", "kubectl pods --name=\"^my-\"")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
