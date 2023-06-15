package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"kubectl-plugin-pod/handlers"
	"sigs.k8s.io/yaml"
)

var podGetByCacheCmd = &cobra.Command{
	Use:    "list-cache",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("pod name is required")
		}

		pod, err := handlers.Fact.Core().V1().Pods().Lister().Pods(currentNS).Get(args[0])
		if err != nil {
			return err
		}

		b, err := yaml.Marshal(pod)
		if err != nil {
			return err
		}

		fmt.Println(string(b))
		return nil
	},
}
