package pod

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	v1 "k8s.io/api/core/v1"
	apilabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/json"
	"kubectl-plugin-pod/config"
	"kubectl-plugin-pod/handlers"
	"kubectl-plugin-pod/tools"
	"log"
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
func getPodDetailByGjson(podName string, item podGetItem) {
	// 获取pod对象
	pod, err := handlers.Factory().Core().V1().Pods().Lister().Pods(currentNS).Get(podName)
	if err != nil {
		log.Println(err)
		return
	}

	if item.path == PodPathLog { // 查看pod日志
		// 代表是sidecar，并且没有选择查看的容器，调用选择容器列表的bubbleTea界面
		if len(pod.Spec.Containers) > 1 && item.containerName == "" {
			// 使用协程，TODO 否则会出现各种bug，原因未知
			// 比如第二个下拉列表需要选择2次才能出来结果，还有按上键2次才能显示上一次的命令等
			go podGetLogBubbleTea(podName, pod.Spec.Containers)
			return
		}

		var tailLine int64 = 100
		req := config.Clientset.CoreV1().Pods(currentNS).
			GetLogs(podName, &v1.PodLogOptions{
				Container: item.containerName, // 指定容器名称
				TailLines: &tailLine,
			})
		ret := req.Do(context.Background())
		b, err := ret.Raw()
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println(string(b))
		return
	}

	if item.path == PodPathEvent { // 查看pod事件
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
		tools.PrintEvent(podEvents)
		return
	}

	b, err := json.Marshal(pod)
	if err != nil {
		log.Println(err)
		return
	}

	ret := gjson.Get(string(b), item.path)
	if !ret.Exists() {
		log.Println("无法获取对应的内容:", item.path)
		return
	}

	if !ret.IsObject() && !ret.IsArray() { // 不是对象或数组直接打印
		fmt.Println(ret.Raw)
		return
	}

	// 把json字符串转为yaml
	var retMap interface{}
	if ret.IsObject() {
		retMap = make(map[string]interface{})
	} else if ret.IsArray() {
		retMap = []interface{}{}
	}
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
