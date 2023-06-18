package deploy

import (
	"context"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"kubectl-plugin-pod/config"
	"kubectl-plugin-pod/handlers"
	"kubectl-plugin-pod/tools"
	"log"
	"sigs.k8s.io/yaml"
	"sort"
	"strings"
	"time"
)

// view组成
type viewComponent struct {
	flex    *tview.Flex
	ns      *tview.DropDown
	depList *tview.List
	detail  *tview.TextView
	podList *tview.List
	footer  *tview.TextView
	ok      bool
}

var viewComp = &viewComponent{}

var viewCmd = &cobra.Command{
	Use:          "view",
	Example:      "kubectl deploy view",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		app := tview.NewApplication()

		viewNs := renderNamespace(app)
		viewDepList := renderDeployView(app)
		viewDetail := renderDetail(app)
		viewPodList := renderPodView(app)
		viewFooter := renderFooter(app)

		// 代表加入组件组成
		viewComp.ns = viewNs
		viewComp.detail = viewDetail
		viewComp.depList = viewDepList
		viewComp.podList = viewPodList
		viewComp.footer = viewFooter

		/*grid := tview.NewGrid().
			SetRows(1, 0, 3).
			SetColumns(30, 0, 30).
			SetBorders(true).
			AddItem(viewNs, 0, 0, 1, 3, 0, 0, false).
			AddItem(viewFooter, 2, 0, 1, 3, 0, 0, false)

		// Layout for screens wider than 100 cells.
		grid.AddItem(viewDepList, 1, 0, 1, 1, 0, 100, true).
			AddItem(viewDetail, 1, 1, 1, 1, 0, 100, false).
			AddItem(viewPodList, 1, 2, 1, 1, 0, 100, false)*/

		flex := tview.NewFlex().
			AddItem(viewDepList, 25, 1, true).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(viewNs, 0, 1, false).
				AddItem(viewDetail, 0, 6, false).
				AddItem(viewFooter, 5, 2, false), 0, 2, false).
			AddItem(viewPodList, 25, 1, false)
		viewComp.flex = flex

		// namespace默认选中第一个: default
		viewNs.SetCurrentOption(0)

		viewComp.ok = true

		// 监听deploy，有变更后重新渲染列表
		go func() {
			for _ = range tools.DeployChan {
				if viewComp.ok {
					flushDeployView(viewDepList, app)
				}
			}
		}()

		// 监听pod，如果正在查看，则更新
		go func() {
			for pod := range tools.PodChan {
				if viewDepList.GetItemCount() > 0 {
					// 获取当前选中的deploy名称
					oldItem, _ := viewDepList.GetItemText(viewDepList.GetCurrentItem())
					depList := getDeploymentsByPod(pod)
					for _, dep := range depList {
						if dep.Name == oldItem {
							viewComp.footer.SetText("new pod set:" + dep.Name + time.Now().Format("15:04:05"))
							// 手动触发deploy选中，重新渲染详情和pod列表
							deployAddItemSelected(dep.Name, app)()
						}
					}
					app.ForceDraw()
				}
			}
		}()

		// 监听键盘事件
		app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyRune { // 是字母
				if l, ok := app.GetFocus().(*tview.List); ok { // 焦点是pod列表
					if strings.Index(l.GetTitle(), "Pods") >= 0 { // 标题包含Pods
						if event.Rune() == 'd' { // 删除pod
							deleteCurrentPod(app)
						}
					}
				}
			}
			return event
		})

		if err := app.SetRoot(flex, true).Run(); err != nil {
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
			viewComp.depList.Clear()
			viewComp.detail.Clear()
			viewComp.podList.Clear()

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

	dropdown.SetBorder(true)

	for _, ns := range nsList {
		if ns.Name != "default" {
			dropdown.AddOption(ns.Name, selected(ns.Name))
		}
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

	list.SetBorder(true)
	list.SetTitle("Deployment列表")

	// 立刻渲染列表
	flushDeployView(list, app)

	// esc切换到详情
	list.SetDoneFunc(func() {
		app.SetFocus(viewComp.ns)
	})

	return list
}

// 重新渲染deploy列表
func flushDeployView(list *tview.List, app *tview.Application) {
	// 查询deploy列表
	depList := listDeploy()

	// 获取当前选中的deploy名称
	var oldItem string
	if list.GetItemCount() > 0 {
		oldItem, _ = list.GetItemText(list.GetCurrentItem())
	}
	newIndex := -1

	if depList != nil {
		list.Clear()
		sort.Sort(sortDeployByName(depList)) // 按首字母排序

		for index, dep := range depList {
			depName := dep.Name
			if oldItem == depName { // 代表新列表的这个选项是之前选中的
				newIndex = index
			}
			list.AddItem(dep.Name, fmt.Sprintf("%d/%d", dep.Status.ReadyReplicas, dep.Status.Replicas), rune(dep.Name[0]), deployAddItemSelected(depName, app))
		}

		list.SetTitle(fmt.Sprintf("deployments(%d)", len(depList)))
	}

	if newIndex == -1 { // 代表没有找到刚刚选中的deploy，可能被删除了，清空yaml详情和pod列表
		newIndex = 0
		if viewComp.ok {
			viewComp.detail.Clear()
			viewComp.podList.Clear()
			app.SetFocus(viewComp.depList)
		}
	}

	list.SetCurrentItem(newIndex) // 重新定位选中项

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

	textView.SetBorder(true)
	textView.SetTitle("Yaml")

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

	list.SetBorder(true)
	list.SetTitle("Pods")

	// esc切换到详情
	list.SetDoneFunc(func() {
		app.SetFocus(viewComp.detail)
	})

	return list
}

// 渲染底部说明
func renderFooter(app *tview.Application) *tview.TextView {
	textView := tview.NewTextView().SetWordWrap(true)

	textView.SetBorder(true)
	textView.SetTitle("events")

	textView.SetBlurFunc(func() {
		textView.SetBackgroundColor(tcell.Color16) // black
	})
	textView.SetFocusFunc(func() {
		textView.SetBackgroundColor(tcell.Color23)
	})

	return textView
}

// 删除选中的pod
func deleteCurrentPod(app *tview.Application) {
	if viewComp.podList.GetItemCount() > 0 {
		modal := tview.NewModal().
			SetText("Confirm deleting the pod").
			AddButtons([]string{"yes", "no"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonLabel == "yes" {
					podName, _ := viewComp.podList.GetItemText(viewComp.podList.GetCurrentItem())
					config.Clientset.CoreV1().Pods(tools.CurrentDeployNS).Delete(context.Background(), podName, metav1.DeleteOptions{})
				}

				// 切回根为flex
				app.SetRoot(viewComp.flex, true)
				app.SetFocus(viewComp.podList)
			})

		// 设置根为该模态框
		app.SetRoot(modal, false)
	}
}
