package deploy

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"kubectl-plugin-pod/suggestions"
	"kubectl-plugin-pod/tools"
	"os"
	"strings"
)

// 交互式窗口当前的namespace
var currentNS = "default"
var myConsoleWriter = prompt.NewStdoutWriter() // 定义一个自己的writer

var promptCmd = &cobra.Command{
	Use:          "prompt",
	Example:      "kubectl deploy prompt",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		p := prompt.New(
			executorCmd(cmd),
			completer,
			prompt.OptionPrefix(">>> "),
			prompt.OptionWriter(myConsoleWriter),
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
			tools.ResetSTTY()
			os.Exit(0)
		case "use":
			if tools.CheckArgsLen(args, 1) {
				currentNS = args[0]
				fmt.Println("切换namespace为:", blocks[1])
			}
		case "clear":
			clearConsole()
		case "list":
			renderDeploy()
		}
	}

}

var cmdSuggestions = []prompt.Suggest{
	// Command
	{"list", "显示deployment列表"},
	{"use", "切换namespace"},
	{"clear", "清除控制台输出"},
	{"exit", "退出交互式窗口"},
}

func completer(in prompt.Document) []prompt.Suggest {
	w := in.GetWordBeforeCursor()
	if w == "" {
		return []prompt.Suggest{}
	}

	// 判断命令，进行自动提示
	cmd, opt := tools.ParseCmd(in.TextBeforeCursor())
	switch cmd {
	case "use":
		return prompt.FilterHasPrefix(suggestions.NamespaceSuggestions, opt, true)
	case "get":
		return prompt.FilterHasPrefix(suggestions.DeploySuggestions, opt, true)
	default:
		return prompt.FilterHasPrefix(cmdSuggestions, w, true)
	}
}

// 清除控制台输出
func clearConsole() {
	myConsoleWriter.EraseScreen()
	myConsoleWriter.CursorGoTo(0, 0)
	myConsoleWriter.Flush()
}
