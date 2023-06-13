package cmd

import (
	"context"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"kubectl-plugin-pod/config"
	"log"
	"os"
)

var clientset *kubernetes.Clientset

func RunCmd() {
	clientset = config.NewK8sConfig().InitClient()
	config.MergeFlags(rootCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

var rootCmd = &cobra.Command{
	Use:          "pods [flags]",
	Example:      "kubectl pods [flags]",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ns, err := cmd.Flags().GetString("namespace")
		if err != nil {
			return err
		}

		if ns == "" {
			ns = "default"
		}
		podList, err := clientset.CoreV1().Pods(ns).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return err
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Namespace", "Status", "IP"})
		for _, pod := range podList.Items {
			table.Append([]string{pod.Name, pod.Namespace, string(pod.Status.Phase), pod.Status.PodIP})
		}
		table.Render()

		return nil
	},
}
