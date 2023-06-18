package handlers

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"kubectl-plugin-pod/tools"
)

type PodHandler struct{}

func (p PodHandler) OnAdd(obj interface{}) {
	addPodFlushChan(obj)
}

func (p PodHandler) OnUpdate(oldObj, newObj interface{}) {
	addPodFlushChan(newObj)
}

func (p PodHandler) OnDelete(obj interface{}) {
	addPodFlushChan(obj)
}

var _ cache.ResourceEventHandler = &PodHandler{}

// 向chan发送数据，代表需要重新渲染pod列表
func addPodFlushChan(obj interface{}) {
	if !syncDone {
		return
	}

	if pod, ok := obj.(*corev1.Pod); ok {
		if tools.CurrentDeployNS == pod.Namespace {
			tools.PodChan <- pod
		}
	}
}
