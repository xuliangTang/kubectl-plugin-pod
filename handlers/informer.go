package handlers

import (
	"k8s.io/client-go/informers"
	"kubectl-plugin-pod/config"
)

var Fact informers.SharedInformerFactory

func InitFact() {
	Fact = informers.NewSharedInformerFactory(config.Clientset, 0)
	Fact.Core().V1().Pods().Informer().AddEventHandler(&PodHandler{})

	ch := make(chan struct{})
	Fact.Start(ch)
	Fact.WaitForCacheSync(ch)
}
