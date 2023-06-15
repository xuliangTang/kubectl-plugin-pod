package suggestions

import (
	"context"
	"github.com/c-bata/go-prompt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubectl-plugin-pod/config"
	"log"
)

// NamespaceSuggestions 命名空间自动提示列表
var NamespaceSuggestions []prompt.Suggest

func init() {
	nsList, err := config.Clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Println(err)
		return
	}

	for _, ns := range nsList.Items {
		NamespaceSuggestions = append(NamespaceSuggestions, prompt.Suggest{
			Text:        ns.Name,
			Description: string(ns.Status.Phase),
		})
	}
}
