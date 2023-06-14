package cmd

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
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
		switch blocks[0] {
		case "exit":
			fmt.Println("Bye!")
			os.Exit(0)
		case "list":
			if err := podListCmd.RunE(cmd, []string{}); err != nil {
				log.Fatalln(err)
			}
		}
	}

}

var suggestions = []prompt.Suggest{
	// Command
	{"test", "this is test"},
	{"exit", "Exit prompt"},
}

func completer(in prompt.Document) []prompt.Suggest {
	w := in.GetWordBeforeCursor()
	if w == "" {
		return []prompt.Suggest{}
	}
	return prompt.FilterHasPrefix(suggestions, w, true)
}
