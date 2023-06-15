package cmd

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"kubectl-plugin-pod/handlers"
	"log"
	"os"
	"strings"
)

var promptCmd = &cobra.Command{
	Use:          "prompt",
	Example:      "kubectl pods prompt",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		p := prompt.New(
			executorCmd(cmd),
			completer,
			prompt.OptionPrefix(">>> "),
		)
		p.Run()
		return nil
	},
}

func executorCmd(cmd *cobra.Command) func(in string) {
	return func(in string) {
		in = strings.TrimSpace(in)
		blocks := strings.Split(in, " ")
		var args []string
		if len(blocks) > 1 {
			args = blocks[1:]
		}

		switch blocks[0] {
		case "exit":
			fmt.Println("Bye!")
			os.Exit(0)
		case "list":
			handlers.InitFact()
			if err := podListByCacheCmd.RunE(cmd, args); err != nil {
				log.Fatalln(err)
			}
		}
	}

}

var suggestions = []prompt.Suggest{
	// Command
	{"list", "显示pod列表"},
	{"exit", "退出交互式窗口"},
}

func completer(in prompt.Document) []prompt.Suggest {
	w := in.GetWordBeforeCursor()
	if w == "" {
		return []prompt.Suggest{}
	}
	return prompt.FilterHasPrefix(suggestions, w, true)
}
