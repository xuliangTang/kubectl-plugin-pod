package tools

import appsv1 "k8s.io/api/apps/v1"
import corev1 "k8s.io/api/core/v1"

var DeployChan = make(chan *appsv1.Deployment)
var PodChan = make(chan *corev1.Pod)
