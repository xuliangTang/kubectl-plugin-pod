package deploy

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
	"kubectl-plugin-pod/handlers"
	"kubectl-plugin-pod/tools"
	"log"
	"os"
	"sort"
)

// 获取deployment列表
func listDeploy() []*appsv1.Deployment {
	depList, err := handlers.Factory().Apps().V1().Deployments().Lister().
		Deployments(currentNS).List(labels.Everything())
	if err != nil {
		log.Println(err)
		return nil
	}

	sort.Sort(sortDeploy(depList))
	return depList
}

// 渲染deployment列表
func renderDeploy() {
	depList := listDeploy()
	if depList == nil {
		return
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Ready", "Container", "CreatedAt"})
	for _, dep := range depList {
		ready := fmt.Sprintf("%d/%d", dep.Status.ReadyReplicas, dep.Status.Replicas)
		createdAt := dep.CreationTimestamp.Format("2006-01-02 15:04")
		lenContainers := fmt.Sprintf("%d", len(dep.Spec.Template.Spec.Containers))
		depRow := []string{dep.Name, ready, lenContainers, createdAt}
		table.Append(depRow)
	}

	tools.SetTable(table)
	table.Render()
}

// 按时间排序
type sortDeploy []*appsv1.Deployment

func (this sortDeploy) Len() int {
	return len(this)
}
func (this sortDeploy) Less(i, j int) bool {
	return this[i].CreationTimestamp.Time.After(this[j].CreationTimestamp.Time)
}
func (this sortDeploy) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

// 按名称首字母排序
type sortDeployByName []*appsv1.Deployment

func (this sortDeployByName) Len() int {
	return len(this)
}
func (this sortDeployByName) Less(i, j int) bool {
	return []rune(this[i].Name)[0] < []rune(this[j].Name)[0]
}
func (this sortDeployByName) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}
