package handlers

import (
	appsv1 "k8s.io/api/apps/v1"
	"kubectl-plugin-pod/tools"
)

type DeployHandler struct{}

func (p DeployHandler) OnAdd(obj interface{}) {
	addDepFlushChan(obj)
}

func (p DeployHandler) OnUpdate(oldObj, newObj interface{}) {
	addDepFlushChan(newObj)
}

func (p DeployHandler) OnDelete(obj interface{}) {
	addDepFlushChan(obj)
}

// 向chan发送数据，代表需要重新渲染左边的deploy列表
func addDepFlushChan(obj interface{}) {
	if !syncDone {
		return
	}

	if dep, ok := obj.(*appsv1.Deployment); ok {
		if tools.CurrentDeployNS == dep.Namespace {
			tools.DeployChan <- dep
		}
	}
}
