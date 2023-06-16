package cmd

import (
	"context"
	"fmt"
	"github.com/olekukonko/tablewriter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubectl-plugin-pod/config"
	"log"
	"os"
)

// 获取Pod的资cpu/内存使用情况
func getPodCapacityUsage() {
	metricList, err := config.MetricsClient.MetricsV1beta1().
		PodMetricses(currentNS).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Println(err)
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	headers := []string{"Name", "CPU(cores)", "Memory(MB)"}
	table.SetHeader(headers)
	var data [][]string
	for _, pod := range metricList.Items {
		for _, c := range pod.Containers {
			cpu := fmt.Sprintf("%dm", c.Usage.Cpu().MilliValue())
			mem := fmt.Sprintf("%dm", c.Usage.Memory().Value()/1024/1024)
			podRow := []string{pod.Name, cpu, mem}
			data = append(data, podRow)
		}
	}
	table.AppendBulk(data)
	table.SetRowLine(true)
	table.SetAutoMergeCellsByColumnIndex([]int{0})
	table.Render()
}
