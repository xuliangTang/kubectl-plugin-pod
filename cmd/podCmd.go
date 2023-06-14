package cmd

import (
	"context"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"kubectl-plugin-pod/config"
	"kubectl-plugin-pod/tools"
	"log"
	"os"
)

var clientset *kubernetes.Clientset

func RunCmd() {
	clientset = config.NewK8sConfig().InitClient()
	config.MergeFlags(podCmd)

	podCmd.Flags().BoolVar(&showLabels, "show-labels", false, "kubectl pods --show-labels")
	podCmd.Flags().StringVar(&labels, "labels", "", "kubectl pods --labels=\"app=test,version=v1\"")

	if err := podCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

var podCmd = &cobra.Command{
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
		podList, err := clientset.CoreV1().Pods(ns).List(context.Background(), metav1.ListOptions{
			LabelSelector: labels,
		})
		if err != nil {
			return err
		}

		table := tablewriter.NewWriter(os.Stdout)
		headers := []string{"Name", "Namespace", "Status", "IP"}
		if showLabels {
			headers = append(headers, "Labels")
		}
		table.SetHeader(headers)

		for _, pod := range podList.Items {
			podRow := []string{pod.Name, pod.Namespace, string(pod.Status.Phase), pod.Status.PodIP}
			if showLabels {
				podRow = append(podRow, tools.Map2String(pod.Labels))
			}
			table.Append(podRow)
		}
		tools.SetTable(table)
		table.Render()

		return nil
	},
}
