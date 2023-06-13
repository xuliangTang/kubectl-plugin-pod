package main

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubectl-plugin-pod/config"
	"kubectl-plugin-pod/tools"
)

func main() {
	clientset := config.NewK8sConfig().InitClient()
	podList, err := clientset.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{})
	tools.Check(err)

	for _, pod := range podList.Items {
		fmt.Println(pod.Name)
	}
}
