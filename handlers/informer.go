package handlers

import (
	"k8s.io/client-go/informers"
	"kubectl-plugin-pod/config"
	"sync"
)

var fact informers.SharedInformerFactory
var once sync.Once

// Factory 初始化informerFactory
func Factory() informers.SharedInformerFactory {
	once.Do(func() {
		fact = informers.NewSharedInformerFactory(config.Clientset, 0)
		fact.Core().V1().Pods().Informer().AddEventHandler(&PodHandler{})
		fact.Core().V1().Events().Informer().AddEventHandler(&EventHandler{})

		ch := make(chan struct{})
		fact.Start(ch)
		fact.WaitForCacheSync(ch)
	})

	return fact
}
