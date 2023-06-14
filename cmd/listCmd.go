package cmd

import (
	"context"
	"encoding/json"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubectl-plugin-pod/tools"
	"os"
	"regexp"
)

var podListCmd = &cobra.Command{
	Use:          "list",
	Example:      "kubectl pods list [flags]",
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
			FieldSelector: fields,
		})
		if err != nil {
			return err
		}

		if err = filterName(podList); err != nil {
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

func filterName(list *v1.PodList) error {
	if name == "" {
		return nil
	}

	b, err := json.Marshal(list)
	if err != nil {
		return err
	}

	ret := gjson.Get(string(b), "items.#.metadata.name")

	var newItems []v1.Pod
	for k, r := range ret.Array() {
		if m, err := regexp.MatchString(name, r.String()); err == nil && m {
			newItems = append(newItems, list.Items[k])
		}
	}
	list.Items = newItems

	return nil
}
