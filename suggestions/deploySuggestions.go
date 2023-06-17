package suggestions

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"k8s.io/apimachinery/pkg/labels"
	"kubectl-plugin-pod/handlers"
	"log"
)

// DeploySuggestions pod列表自动提示
var DeploySuggestions []prompt.Suggest

func init() {
	depList, err := handlers.Factory().Apps().V1().Deployments().Lister().List(labels.Everything())
	if err != nil {
		log.Println(err)
		return
	}

	for _, dep := range depList {
		DeploySuggestions = append(DeploySuggestions, prompt.Suggest{
			Text:        dep.Name,
			Description: fmt.Sprintf("Ready:%d/%d", dep.Status.ReadyReplicas, dep.Status.Replicas),
		})
	}
}
