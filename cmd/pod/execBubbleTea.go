package pod

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"io"
	"kubectl-plugin-pod/handlers"
	"log"
	"os"
	"strings"
)

// exec pod shell选择容器的界面
func execBubbleTea(podName string) {
	// 获取pod对象
	pod, err := handlers.Factory().Core().V1().Pods().Lister().Pods(currentNS).Get(podName)
	if err != nil {
		log.Println(err)
	}

	if len(pod.Spec.Containers) == 1 { // 只有一个容器，无需选择
		podContainerTerminal(podName, "")
		return
	}

	var items []list.Item
	for _, p := range pod.Spec.Containers {
		items = append(items, podExecItem(p.Name))
	}

	m := newPodExecModel(items, podName)

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

type podExecItemDelegate struct{}

func (d podExecItemDelegate) Height() int                             { return 1 }
func (d podExecItemDelegate) Spacing() int                            { return 0 }
func (d podExecItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d podExecItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(podExecItem)
	if !ok {
		return
	}

	// 输出选项内容
	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type podExecItem string

func (i podExecItem) FilterValue() string { return "" }

type podExecModel struct {
	podName  string
	list     list.Model
	choice   string
	quitting bool
}

func newPodExecModel(items []list.Item, podName string) *podExecModel {
	const defaultWidth = 20
	const listHeight = 14
	l := list.New(items, podExecItemDelegate{}, defaultWidth, listHeight)
	l.Title = "请选择容器"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return &podExecModel{list: l, podName: podName}
}

func (m podExecModel) Init() tea.Cmd {
	return nil
}

func (m podExecModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			i, ok := m.list.SelectedItem().(podExecItem)
			if ok {
				m.choice = string(i)
			}

			// 进入pod shell
			podContainerTerminal(m.podName, string(i))

			// 退出界面
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m podExecModel) View() string {
	return "\n" + m.list.View()
}
