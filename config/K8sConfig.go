package config

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"kubectl-plugin-pod/tools"
)

var cfgFlags *genericclioptions.ConfigFlags
var Clientset *kubernetes.Clientset

type K8sConfig struct{}

func NewK8sConfig() *K8sConfig {
	return &K8sConfig{}
}

// K8sRestConfigFromCli 获取restConfig（手动指定kubeconfig文件路径）
func (*K8sConfig) K8sRestConfigFromCli() *rest.Config {
	cfgFlags = genericclioptions.NewConfigFlags(true)
	config, err := cfgFlags.ToRawKubeConfigLoader().ClientConfig()
	tools.Check(err)

	return config
}

// InitClient 创建clientset
func (this *K8sConfig) InitClient() *kubernetes.Clientset {
	c, err := kubernetes.NewForConfig(this.K8sRestConfigFromCli())
	tools.Check(err)
	return c
}

// InitDynamicClient 创建dynamicClient
func (this *K8sConfig) InitDynamicClient() dynamic.Interface {
	client, err := dynamic.NewForConfig(this.K8sRestConfigFromCli())
	tools.Check(err)
	return client
}

func MergeFlags(cmds ...*cobra.Command) {
	for _, cmd := range cmds {
		cfgFlags.AddFlags(cmd.Flags())
	}
}
