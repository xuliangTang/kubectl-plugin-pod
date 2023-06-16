package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	v1 "k8s.io/api/core/v1"
	apilabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/json"
	"kubectl-plugin-pod/config"
	"kubectl-plugin-pod/handlers"
	"kubectl-plugin-pod/tools"
	"log"
	"os"
	"sigs.k8s.io/yaml"
)

var podGetByCacheCmd = &cobra.Command{
	Use:    "list-cache",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("pod name is required")
		}

		pod, err := handlers.Factory().Core().V1().Pods().Lister().Pods(currentNS).Get(args[0])
		if err != nil {
			return err
		}

		b, err := yaml.Marshal(pod)
		if err != nil {
			return err
		}

		fmt.Println(string(b))
		return nil
	},
}

// 根据gjson的规则获取pod的指定内容
func getPodDetailByGjson(podName, path string) {
	// 获取pod对象
	pod, err := handlers.Factory().Core().V1().Pods().Lister().Pods(currentNS).Get(podName)
	if err != nil {
		log.Println(err)
		return
	}

	if path == PodPathLog { // 查看pod日志
		req := config.Clientset.CoreV1().Pods(currentNS).
			GetLogs(podName, &v1.PodLogOptions{})
		ret := req.Do(context.Background())
		b, err := ret.Raw()
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println(string(b))
		return
	}

	if path == PodPathEvent { // 查看pod事件
		eventList, err := handlers.Factory().Core().V1().Events().Lister().List(apilabels.Everything())
		if err != nil {
			log.Println(err)
			return
		}

		var podEvents []*v1.Event
		for _, e := range eventList {
			// 对比event的所属资源是否是当前pod
			if e.InvolvedObject.UID == pod.UID {
				podEvents = append(podEvents, e)
			}
		}
		printEvent(podEvents)
		return
	}

	b, err := json.Marshal(pod)
	if err != nil {
		log.Println(err)
		return
	}

	ret := gjson.Get(string(b), path)
	if !ret.Exists() {
		log.Println("无法获取对应的内容:", path)
		return
	}

	if !ret.IsObject() && !ret.IsArray() { // 不是对象或数组直接打印
		fmt.Println(ret.Raw)
		return
	}

	// 把json字符串转为yaml
	retMap := make(map[string]interface{})
	err = yaml.Unmarshal([]byte(ret.Raw), &retMap)
	if err != nil {
		log.Println(err)
		return
	}
	retYaml, err := yaml.Marshal(retMap)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(string(retYaml))
}

// 事件要显示的头
var eventHeaders = []string{"Type", "Reason", "Object", "Message"}

func printEvent(events []*v1.Event) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(eventHeaders)
	for _, e := range events {
		podRow := []string{e.Type, e.Reason,
			fmt.Sprintf("%s/%s", e.InvolvedObject.Kind, e.InvolvedObject.Name), e.Message}

		table.Append(podRow)
	}
	tools.SetTable(table)
	table.Render()
}
