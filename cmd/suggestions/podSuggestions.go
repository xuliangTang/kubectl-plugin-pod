package suggestions

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	apilabels "k8s.io/apimachinery/pkg/labels"
	"kubectl-plugin-pod/handlers"
	"log"
)

// PodSuggestions pod列表自动提示
var PodSuggestions []prompt.Suggest

func init() {
	podList, err := handlers.Factory().Core().V1().Pods().Lister().List(apilabels.Everything())
	if err != nil {
		log.Println(err)
		return
	}

	for _, pod := range podList {
		PodSuggestions = append(PodSuggestions, prompt.Suggest{
			Text:        pod.Name,
			Description: fmt.Sprintf("%s / %s / %s", pod.Namespace, pod.Status.Phase, pod.Spec.NodeName),
		})
	}
}
