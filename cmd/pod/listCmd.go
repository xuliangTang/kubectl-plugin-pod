package pod

import (
	"context"
	"encoding/json"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apilabels "k8s.io/apimachinery/pkg/labels"
	"kubectl-plugin-pod/config"
	"kubectl-plugin-pod/handlers"
	"kubectl-plugin-pod/tools"
	"os"
	"regexp"
	"sort"
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
		podList, err := config.Clientset.CoreV1().Pods(ns).List(context.Background(), metav1.ListOptions{
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

var podListByCacheCmd = &cobra.Command{
	Use:    "list-cache",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		podList, err := handlers.Factory().Core().V1().Pods().Lister().Pods(currentNS).List(apilabels.Everything())
		if err != nil {
			return err
		}

		sort.Sort(coreV1PodList(podList))

		table := tablewriter.NewWriter(os.Stdout)
		headers := []string{"Name", "Namespace", "Status", "IP"}
		if showLabels {
			headers = append(headers, "Labels")
		}
		table.SetHeader(headers)

		for _, pod := range podList {
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

// 按时间排序
type coreV1PodList []*v1.Pod

func (this coreV1PodList) Len() int {
	return len(this)
}
func (this coreV1PodList) Less(i, j int) bool {
	return this[i].CreationTimestamp.Time.After(this[j].CreationTimestamp.Time)
}
func (this coreV1PodList) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

func addListFlags() {
	podListCmd.Flags().BoolVar(&showLabels, "show-labels", false, "kubectl pods --show-labels")
	podListCmd.Flags().StringVar(&labels, "labels", "", "kubectl pods --labels=\"app=test,version=v1\"")
	podListCmd.Flags().StringVar(&fields, "fields", "", "kubectl pods --fields=\"status.phase=Running\"")
	podListCmd.Flags().StringVar(&name, "name", "", "kubectl pods --name=\"^my-\"")
}
