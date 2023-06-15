package cmd

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
	"os"
	"strings"
)

// 运行获取pod详情的界面
func podGetBubbleTea(podName string) {
	items := []list.Item{
		podGetItem{title: "查看全部", path: "@this"},
		podGetItem{title: "查看metadata", path: "metadata"},
		podGetItem{title: "查看spec", path: "spec"},
		podGetItem{title: "查看labels", path: "metadata.labels"},
		podGetItem{title: "查看annotations", path: "metadata.annotations"},
	}

	const defaultWidth = 20
	const listHeight = 14
	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "请选择你的操作"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := podGetModel{list: l, podName: podName}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(podGetItem)
	if !ok {
		return
	}

	// 输出选项内容
	str := fmt.Sprintf("%d. %s", index+1, i.title)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type podGetItem struct {
	title string // 选项名称
	path  string // gjson过滤规则
}

func (i podGetItem) FilterValue() string { return "" }

type podGetModel struct {
	podName  string
	list     list.Model
	choice   string
	quitting bool
}

func (m podGetModel) Init() tea.Cmd {
	return nil
}

func (m podGetModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter": // 代表已选择
			i, ok := m.list.SelectedItem().(podGetItem)
			if ok {
				m.choice = i.title
			}

			// 获取并输出pod指定的内容
			getPodDetailByGjson(m.podName, i.path)
			// 退出界面
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m podGetModel) View() string {
	if m.quitting {
		return quitTextStyle.Render("已退出")
	}
	return "\n" + m.list.View()
}
