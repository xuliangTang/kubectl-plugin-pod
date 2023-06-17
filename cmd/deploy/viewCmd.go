package deploy

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
	"kubectl-plugin-pod/handlers"
	"sigs.k8s.io/yaml"
	"sort"
)

// view组成
type viewComponent struct {
	depList *tview.List
	detail  *tview.TextView
}

var viewComp = &viewComponent{}

var viewCmd = &cobra.Command{
	Use:          "view",
	Example:      "kubectl deploy view",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		app := tview.NewApplication()

		newPrimitive := func(text string) tview.Primitive {
			return tview.NewTextView().
				SetTextAlign(tview.AlignCenter).
				SetText(text)
		}
		list := renderDeployView(app)
		detail := renderDetail(app)
		pod := newPrimitive("Pods")

		// 代表加入组件组成
		viewComp.detail = detail
		viewComp.depList = list

		grid := tview.NewGrid().
			SetRows(3, 0, 3).
			SetColumns(30, 0, 30).
			SetBorders(true).
			AddItem(newPrimitive("Header"), 0, 0, 1, 3, 0, 0, false).
			AddItem(newPrimitive("事件和说明"), 2, 0, 1, 3, 0, 0, false)

		// Layout for screens wider than 100 cells.
		grid.AddItem(list, 1, 0, 1, 1, 0, 100, true).
			AddItem(detail, 1, 1, 1, 1, 0, 100, false).
			AddItem(pod, 1, 2, 1, 1, 0, 100, false)

		if err := app.SetRoot(grid, true).EnableMouse(true).Run(); err != nil {
			panic(err)
		}

		return nil
	},
}

// 渲染deployment view界面
func renderDeployView(app *tview.Application) *tview.List {
	depList := listDeploy()
	if depList == nil {
		return nil
	}
	sort.Sort(sortDeployByName(depList)) // 按首字母排序

	// 插入列表
	list := tview.NewList()
	list.SetBlurFunc(func() {
		list.SetBackgroundColor(tcell.Color16) // black
	})
	list.SetFocusFunc(func() {
		list.SetBackgroundColor(tcell.Color23)
	})
	for _, dep := range depList {
		depName := dep.Name
		list.AddItem(dep.Name, "", rune(dep.Name[0]), func() {
			viewComp.detail.SetText("")
			getDep, err := handlers.Factory().Apps().V1().Deployments().Lister().Deployments(currentNS).Get(depName)
			if err != nil {
				viewComp.detail.SetText(err.Error())
				return
			}
			b, _ := yaml.Marshal(getDep)
			viewComp.detail.SetText(string(b))
			// 切换焦点到中间详情部分
			app.SetFocus(viewComp.detail)
		})
	}
	list.AddItem("Quit", "Press to exit", 'q', func() {
		app.Stop()
	})

	return list
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
		}
	})

	return textView
}
