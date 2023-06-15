package cmd

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"kubectl-plugin-pod/handlers"
	"log"
	"os"
	"regexp"
	"strings"
)

var promptCmd = &cobra.Command{
	Use:          "prompt",
	Example:      "kubectl pods prompt",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 初始化informerFactory
		handlers.InitFact()
		// 初始化pod列表提示
		if err := initPodSuggestions(); err != nil {
			return err
		}

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
			if err := podListByCacheCmd.RunE(cmd, args); err != nil {
				log.Fatalln(err)
			}
		case "get":

		}
	}

}

var suggestions = []prompt.Suggest{
	// Command
	{"list", "显示pod列表"},
	{"get", "查看pod详情"},
	{"exit", "退出交互式窗口"},
}

func completer(in prompt.Document) []prompt.Suggest {
	w := in.GetWordBeforeCursor()
	if w == "" {
		return []prompt.Suggest{}
	}

	// 判断命令是get，进行自动提示pod列表
	cmd, opt := parseCmd(in.TextBeforeCursor())
	if cmd == "get" {
		return prompt.FilterHasPrefix(podSuggestions, opt, true)
	}

	return prompt.FilterHasPrefix(suggestions, w, true)
}

// 解析字符串的命令和参数
func parseCmd(w string) (string, string) {
	w = regexp.MustCompile("\\s+").ReplaceAllString(w, " ")
	l := strings.Split(w, " ")

	if len(l) >= 2 {
		return l[0], strings.Join(l[1:], " ")
	}
	return w, ""
}
