package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"k8s.io/apimachinery/pkg/util/json"
	"kubectl-plugin-pod/handlers"
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
func getPodDetailByGjson(podName, path string) {
	// 获取pod对象
	pod, err := handlers.Factory().Core().V1().Pods().Lister().Pods(currentNS).Get(podName)
	if err != nil {
		log.Println(err)
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

	// 把json字符串转为yaml
	retMap := make(map[string]interface{})
	err = yaml.Unmarshal([]byte(ret.Raw), &retMap)
	if err != nil {
		log.Println(err)
	}
	retYaml, err := yaml.Marshal(retMap)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(string(retYaml))
}
