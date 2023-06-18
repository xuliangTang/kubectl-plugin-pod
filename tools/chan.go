package tools

import appsv1 "k8s.io/api/apps/v1"

var DeployChan = make(chan *appsv1.Deployment)
