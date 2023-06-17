package deploy

import (
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubectl-plugin-pod/config"
	"log"
	"regexp"
	"strconv"
)

// 伸缩deploy副本
func scaleDeploy(deployName string) {
	p := tea.NewProgram(initialModel(deployName))
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type (
	errMsg error
)

type scaleModel struct {
	deployName string
	textInput  textinput.Model
	err        error
}

func initialModel(deployName string) scaleModel {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return scaleModel{
		deployName: deployName,
		textInput:  ti,
		err:        nil,
	}
}

func (m scaleModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m scaleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			numStr := m.textInput.Value()
			if !checkScale(numStr) {
				log.Println("scale的值必须是0-50之间")
				return m, nil
			}

			num, _ := strconv.Atoi(numStr)
			scale, err := config.Clientset.AppsV1().Deployments(currentNS).GetScale(context.Background(), m.deployName, metav1.GetOptions{})
			if err != nil {
				log.Println(err)
				return m, tea.Quit
			}
			scale.Spec.Replicas = int32(num)
			_, err = config.Clientset.AppsV1().Deployments(currentNS).UpdateScale(context.Background(), m.deployName, scale, metav1.UpdateOptions{})
			if err != nil {
				log.Println(err)
				return m, tea.Quit
			}

			fmt.Println("副本伸缩成功")
			return m, tea.Quit
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m scaleModel) View() string {
	return fmt.Sprintf(
		"请填写伸缩的副本数(0-50之间)\n\n%s\n\n%s",
		m.textInput.View(),
		"(esc to quit)",
	) + "\n"
}

// 正则判断必须是0到50
func checkScale(v string) bool {
	if regexp.MustCompile("^(?:[0-9]|[1-4][0-9]|50)$").MatchString(v) {
		return true
	}
	return false
}
