package deploy

import (
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
	"sort"
)

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
		detail := newPrimitive("详情")
		pod := newPrimitive("Pods")

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
	for _, dep := range depList {
		list.AddItem(dep.Name, "", rune(dep.Name[0]), nil)
	}
	list.AddItem("Quit", "Press to exit", 'q', func() {
		app.Stop()
	})

	return list
}
