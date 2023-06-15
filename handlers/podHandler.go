package handlers

import "k8s.io/client-go/tools/cache"

type PodHandler struct{}

func (p PodHandler) OnAdd(obj interface{}) {
}

func (p PodHandler) OnUpdate(oldObj, newObj interface{}) {
}

func (p PodHandler) OnDelete(obj interface{}) {
}

var _ cache.ResourceEventHandler = &PodHandler{}
