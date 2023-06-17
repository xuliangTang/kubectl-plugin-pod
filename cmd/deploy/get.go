package deploy

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tidwall/gjson"
	"io"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/json"
	"kubectl-plugin-pod/handlers"
	"kubectl-plugin-pod/tools"
	"log"
	"os"
	"sigs.k8s.io/yaml"
	"strings"
)

const (
	DeployPathEvent = "__event__"
	DeployPodsEvent = "__pods__"
)

// 调用查看deploy详情界面
func getDeploy(deployName string) {
	items := []list.Item{
		deployGetItem{title: "查看全部", path: "@this"},
		deployGetItem{title: "查看metadata", path: "metadata"},
		deployGetItem{title: "查看spec", path: "spec"},
		deployGetItem{title: "查看pod列表", path: DeployPodsEvent},
		deployGetItem{title: "查看事件", path: DeployPathEvent},
		deployGetItem{title: "查看labels", path: "metadata.labels"},
		deployGetItem{title: "查看annotations", path: "metadata.annotations"},
	}

	m := newDeployGetModel(items, deployName)

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

// 根据gjson选择器过滤deploy详情
func getDeployDetailByGjson(deployName string, item deployGetItem) {
	dep, err := handlers.Factory().Apps().V1().Deployments().Lister().Deployments(currentNS).Get(deployName)
	if err != nil {
		log.Println(err)
		return
	}

	if item.path == DeployPodsEvent { // 查看pod列表
		podList := getPodsByDeploy(dep)
		tools.PrintPods(podList)
		return
	}

	if item.path == DeployPathEvent { // 查看事件
		eventList, err := handlers.Factory().Core().V1().Events().Lister().Events(currentNS).List(labels.Everything())
		if err != nil {
			log.Println(err)
			return
		}

		var deployEvents []*v1.Event
		for _, e := range eventList {
			if e.InvolvedObject.UID == dep.UID {
				deployEvents = append(deployEvents, e)
			}
		}

		tools.PrintEvent(deployEvents)
		return
	}

	jsonStr, err := json.Marshal(dep)
	if err != nil {
		log.Println(err)
		return
	}

	ret := gjson.Get(string(jsonStr), item.path)
	if !ret.Exists() {
		log.Println("无法获取对应的内容:", item.path)
		return
	}

	if !ret.IsObject() && !ret.IsArray() { // 不是对象或数组直接打印
		fmt.Println(ret.Raw)
		return
	}

	// 把json字符串转为yaml
	var retMap interface{}
	if ret.IsObject() {
		retMap = make(map[string]interface{})
	} else if ret.IsArray() {
		retMap = []interface{}{}
	}
	err = yaml.Unmarshal([]byte(ret.Raw), &retMap)
	if err != nil {
		log.Println(err)
		return
	}
	retYaml, err := yaml.Marshal(retMap)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(string(retYaml))
}

const revision = "deployment.kubernetes.io/revision"

// 获取deploy下的pod列表
func getPodsByDeploy(deploy *appsv1.Deployment) (pods []*v1.Pod) {
	// 获取所有rs
	rsList, err := handlers.Factory().Apps().V1().ReplicaSets().Lister().ReplicaSets(currentNS).List(labels.Everything())
	if err != nil {
		log.Println(err)
		return
	}

	for _, rs := range rsList { // 判断rs是否属于当前deployment
		if rs.Annotations[revision] != deploy.Annotations[revision] {
			continue
		}

		for _, ref := range rs.OwnerReferences {
			if ref.UID == deploy.UID {
				pods = append(pods, getPodsByRs(rs)...)
				break
			}
		}
	}

	return
}

// 获取replicaSet关联的pod
func getPodsByRs(rs *appsv1.ReplicaSet) (pods []*v1.Pod) {
	// 获取所有pod
	podList, err := handlers.Factory().Core().V1().Pods().Lister().Pods(currentNS).List(labels.Everything())
	if err != nil {
		log.Println(err)
		return
	}

	for _, pod := range podList { // 判断pod是否属于当前rs
		for _, ref := range pod.OwnerReferences {
			if ref.UID == rs.UID {
				pods = append(pods, pod)
				break
			}
		}
	}

	return
}

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
)

type deployGetItemDelegate struct{}

func (d deployGetItemDelegate) Height() int                             { return 1 }
func (d deployGetItemDelegate) Spacing() int                            { return 0 }
func (d deployGetItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d deployGetItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(deployGetItem)
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

type deployGetItem struct {
	title string // 选项名称
	path  string // gjson过滤规则
}

func (i deployGetItem) FilterValue() string { return "" }

type deployGetModel struct {
	deployName string
	list       list.Model
	choice     string
	quitting   bool
}

func newDeployGetModel(items []list.Item, deployName string) *deployGetModel {
	const defaultWidth = 20
	const listHeight = 14
	l := list.New(items, deployGetItemDelegate{}, defaultWidth, listHeight)
	l.Title = "请选择你的操作"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return &deployGetModel{list: l, deployName: deployName}
}

func (m deployGetModel) Init() tea.Cmd {
	return nil
}

func (m deployGetModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			i, ok := m.list.SelectedItem().(deployGetItem)
			if ok {
				m.choice = i.title
			}

			// 获取并输出pod指定的内容
			tea.ClearScrollArea()
			getDeployDetailByGjson(m.deployName, i)

			// 退出界面
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m deployGetModel) View() string {
	return "\n" + m.list.View()
}
