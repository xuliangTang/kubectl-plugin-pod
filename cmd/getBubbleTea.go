package cmd

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
	v1 "k8s.io/api/core/v1"
	"os"
	"strings"
)

const (
	PodPathEvent = "__event__"
	PodPathLog   = "__log__"
)

// 运行获取pod详情的界面
func podGetBubbleTea(podName string) {
	items := []list.Item{
		podGetItem{title: "查看全部", path: "@this"},
		podGetItem{title: "查看metadata", path: "metadata"},
		podGetItem{title: "查看spec", path: "spec"},
		podGetItem{title: "查看日志", path: PodPathLog},
		podGetItem{title: "查看事件", path: PodPathEvent},
		podGetItem{title: "查看labels", path: "metadata.labels"},
		podGetItem{title: "查看annotations", path: "metadata.annotations"},
	}

	m := newPodGetModel(items, podName)

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

// 运行选择pod容器获取日志的界面
func podGetLogBubbleTea(podName string, containers []v1.Container) {
	items := make([]list.Item, len(containers))
	for i, c := range containers {
		items[i] = podGetItem{title: c.Name, path: PodPathLog, containerName: c.Name}
	}

	m := newPodGetModel(items, podName)
	m.nextLog = true
	m.list.Title = "请选择查看的容器"

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
	// quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type podGetItemDelegate struct{}

func (d podGetItemDelegate) Height() int                             { return 1 }
func (d podGetItemDelegate) Spacing() int                            { return 0 }
func (d podGetItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d podGetItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
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
	title         string // 选项名称
	path          string // gjson过滤规则
	containerName string // 查看日志的容器名称
}

func (i podGetItem) FilterValue() string { return "" }

type podGetModel struct {
	nextLog  bool // 是否嵌套选择容器界面
	podName  string
	list     list.Model
	choice   string
	quitting bool
}

func newPodGetModel(items []list.Item, podName string) *podGetModel {
	const defaultWidth = 20
	const listHeight = 14
	l := list.New(items, podGetItemDelegate{}, defaultWidth, listHeight)
	l.Title = "请选择你的操作"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return &podGetModel{list: l, podName: podName}
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
			tea.ClearScrollArea()
			// tea.ClearScreen()
			getPodDetailByGjson(m.podName, i)

			// 退出界面
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m podGetModel) View() string {
	return "\n" + m.list.View()
}
