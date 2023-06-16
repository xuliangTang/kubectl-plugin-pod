package pod

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"kubectl-plugin-pod/config"
	"log"
	"os"
)

// 进入pod shell
func podContainerTerminal(podName, containerName string) {
	// 初始化一个Executor，用于与pod容器终端建立长连接
	option := &v1.PodExecOptions{
		Container: containerName,
		Command:   []string{"sh"},
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}

	req := config.Clientset.CoreV1().RESTClient().Post().Resource("pods").
		Namespace(currentNS).
		Name(podName).
		SubResource("exec").
		VersionedParams(option, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(config.ClientsetConfig, "POST", req.URL())
	if err != nil {
		log.Println(exec)
		return
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Tty:    true,
	})

}
