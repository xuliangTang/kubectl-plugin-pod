package tools

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"log"
	"os"
	"os/exec"
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

func ResetSTTY() {
	cc := exec.Command("stty", "-F", "/dev/tty", "echo")
	cc.Stdout = os.Stdout
	cc.Stderr = os.Stderr
	if err := cc.Run(); err != nil {
		log.Println(err)
	}
}
