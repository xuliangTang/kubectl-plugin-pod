package tools

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	v1 "k8s.io/api/core/v1"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// SetTable 设置table的样式
func SetTable(table *tablewriter.Table) {
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t") // pad with tabs
	table.SetNoWhiteSpace(true)
}

// Map2String map连接为字符串
func Map2String(m map[string]string) (ret string) {
	list := make([]string, 0)
	for k, v := range m {
		list = append(list, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.Join(list, ",")
}

// CheckArgsLen 检查args数量
func CheckArgsLen(args []string, l int) (ok bool) {
	if len(args) < l {
		log.Println("missing args")
		return false
	}
	return true
}

// ParseCmd 解析字符串的命令和参数
func ParseCmd(w string) (string, string) {
	w = regexp.MustCompile("\\s+").ReplaceAllString(w, " ")
	l := strings.Split(w, " ")

	if len(l) >= 2 {
		return l[0], strings.Join(l[1:], " ")
	}
	return w, ""
}

// PrintEvent 输出event列表
func PrintEvent(events []*v1.Event) {
	var eventHeaders = []string{"Type", "Reason", "Object", "Message"}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(eventHeaders)
	for _, e := range events {
		podRow := []string{e.Type, e.Reason,
			fmt.Sprintf("%s/%s", e.InvolvedObject.Kind, e.InvolvedObject.Name), e.Message}

		table.Append(podRow)
	}
	SetTable(table)
	table.Render()
}

func ResetSTTY() {
	cc := exec.Command("stty", "-F", "/dev/tty", "echo")
	cc.Stdout = os.Stdout
	cc.Stderr = os.Stderr
	if err := cc.Run(); err != nil {
		log.Println(err)
	}
}

func InArray(arr []string, item string) bool {
	for _, p := range arr {
		if p == item {
			return true
		}
	}
	return false
}
