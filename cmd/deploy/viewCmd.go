package deploy

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/labels"
	"kubectl-plugin-pod/handlers"
	"kubectl-plugin-pod/tools"
	"log"
	"sigs.k8s.io/yaml"
	"sort"
	"time"
)

// view组成
type viewComponent struct {
	ns      *tview.DropDown
	depList *tview.List
	detail  *tview.TextView
	podList *tview.List
	footer  *tview.TextView
}

var viewComp = &viewComponent{}

var viewCmd = &cobra.Command{
	Use:          "view",
	Example:      "kubectl deploy view",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		app := tview.NewApplication()

		ns := renderNamespace(app)
		list := renderDeployView(app)
		detail := renderDetail(app)
		pod := renderPodView(app)
		footer := renderFooter(app)

		// 代表加入组件组成
		viewComp.ns = ns
		viewComp.detail = detail
		viewComp.depList = list
		viewComp.podList = pod
		viewComp.footer = footer

		grid := tview.NewGrid().
			SetRows(1, 0, 3).
			SetColumns(30, 0, 30).
			SetBorders(true).
			AddItem(ns, 0, 0, 1, 3, 0, 0, false).
			AddItem(footer, 2, 0, 1, 3, 0, 0, false)

		// Layout for screens wider than 100 cells.
		grid.AddItem(list, 1, 0, 1, 1, 0, 100, true).
			AddItem(detail, 1, 1, 1, 1, 0, 100, false).
			AddItem(pod, 1, 2, 1, 1, 0, 100, false)

		// namespace默认选中第一个: default
		ns.SetCurrentOption(0)

		if err := app.SetRoot(grid, true).EnableMouse(true).Run(); err != nil {
			panic(err)
		}

		return nil
	},
}

// 渲染顶部namespace
func renderNamespace(app *tview.Application) *tview.DropDown {
	nsList, err := handlers.Factory().Core().V1().Namespaces().Lister().List(labels.Everything())
	if err != nil {
		log.Fatalln(err)
	}

	// 选中后的事件
	selected := func(ns string) func() {
		return func() {
			// 清空其他模块
			if viewComp.depList != nil {
				viewComp.depList.Clear()
			}
			if viewComp.detail != nil {
				viewComp.detail.SetText("")
			}
			if viewComp.podList != nil {
				viewComp.podList.Clear()
			}

			// 切换命名空间
			tools.CurrentDeployNS = ns
			// 重新赋值deploy列表
			*viewComp.depList = *renderDeployView(app)

			// 切换焦点到deploy列表
			app.SetFocus(viewComp.depList)
		}
	}

	// 添加namespace下拉选项
	dropdown := tview.NewDropDown().SetLabel("Please select a namespace (hit Enter): ").
		AddOption("default", selected("default"))

	for _, ns := range nsList {
		dropdown.AddOption(ns.Name, selected(ns.Name))
	}

	// 监听回车切换焦点到deploy列表
	dropdown.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			app.SetFocus(viewComp.depList)
		}
	})

	return dropdown
}

// 渲染deployment列表
func renderDeployView(app *tview.Application) *tview.List {
	// 插入列表
	list := tview.NewList()
	list.SetBlurFunc(func() {
		list.SetBackgroundColor(tcell.Color16) // black
	})
	list.SetFocusFunc(func() {
		list.SetBackgroundColor(tcell.Color23)
	})

	// 立刻渲染列表
	flushDeployView(list, app)

	// 监听deploy，有变更后重新渲染列表
	go func() {
		for _ = range tools.DeployChan {
			viewComp.footer.SetText("监听到了新变更，重新渲染 " + time.Now().Format("15:04:05"))
			flushDeployView(list, app)
		}
	}()

	// esc切换到详情
	list.SetDoneFunc(func() {
		app.SetFocus(viewComp.ns)
	})

	return list
}

// 重新渲染deploy列表
func flushDeployView(list *tview.List, app *tview.Application) {
	// 查询deploy列表
	list.Clear()
	depList := listDeploy()
	if depList != nil {
		sort.Sort(sortDeployByName(depList)) // 按首字母排序

		for _, dep := range depList {
			depName := dep.Name
			list.AddItem(dep.Name, fmt.Sprintf("%d/%d", dep.Status.ReadyReplicas, dep.Status.Replicas), rune(dep.Name[0]), deployAddItemSelected(depName, app))
		}
	}

	// 退出选项
	list.AddItem("Quit", "Press to exit", 'q', func() {
		app.Stop()
	})

	// 强制重新渲染，否则要等待激活焦点后才会渲染
	app.ForceDraw()
}

// deploy列表选中后的回调
func deployAddItemSelected(depName string, app *tview.Application) func() {
	return func() {
		// 选中后设置deploy详情
		viewComp.detail.SetText("")
		getDep, err := handlers.Factory().Apps().V1().Deployments().Lister().Deployments(tools.CurrentDeployNS).Get(depName)
		if err != nil {
			viewComp.detail.SetText(err.Error())
			return
		}
		b, _ := yaml.Marshal(getDep)
		viewComp.detail.SetText(string(b))

		// 设置pod列表
		viewComp.podList.Clear()
		podList := getPodsByDeploy(getDep)
		for _, pod := range podList {
			podName := pod.Name
			viewComp.podList.AddItem(podName, fmt.Sprintf("%s/%s", pod.Spec.NodeName, pod.Status.Phase), []rune(podName)[0], nil)
		}

		// 切换焦点到中间详情部分
		app.SetFocus(viewComp.detail)
	}
}

// 渲染中间部分的详情
func renderDetail(app *tview.Application) *tview.TextView {
	textView := tview.NewTextView().SetWordWrap(true)
	textView.SetBlurFunc(func() {
		textView.SetBackgroundColor(tcell.Color16) // black
	})
	textView.SetFocusFunc(func() {
		textView.SetBackgroundColor(tcell.Color23)
	})

	// 监听键盘事件
	textView.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyESC {
			// 切换焦点到左边的deploy列表
			app.SetFocus(viewComp.depList)
		} else if key == tcell.KeyEnter {
			// 切换焦点到右边的pod列表
			app.SetFocus(viewComp.podList)
		}
	})

	return textView
}

// 渲染右边的pod列表
func renderPodView(app *tview.Application) *tview.List {
	list := tview.NewList()
	list.SetBlurFunc(func() {
		list.SetBackgroundColor(tcell.Color16) // black
	})
	list.SetFocusFunc(func() {
		list.SetBackgroundColor(tcell.Color23)
	})

	// esc切换到详情
	list.SetDoneFunc(func() {
		app.SetFocus(viewComp.detail)
	})

	return list
}

// 渲染底部说明
func renderFooter(app *tview.Application) *tview.TextView {
	textView := tview.NewTextView().SetWordWrap(true)
	textView.SetBlurFunc(func() {
		textView.SetBackgroundColor(tcell.Color16) // black
	})
	textView.SetFocusFunc(func() {
		textView.SetBackgroundColor(tcell.Color23)
	})

	return textView
}
