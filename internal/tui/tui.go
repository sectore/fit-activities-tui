package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	importPath   string
	files        []string
	selectedFile string
}

func InitialModel(path string, files []string) Model {
	return Model{
		importPath:   path,
		files:        files,
		selectedFile: "",
	}
}

func (a Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:

		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		}

	}

	return m, nil
}

func (m Model) View() string {

	s := fmt.Sprintf("path: %s\n", m.importPath)

	var h = lipgloss.NewStyle().
		Border(lipgloss.ThickBorder(), true).PaddingLeft(2).PaddingRight(2).MarginBottom(1)

	s += h.Render(fmt.Sprintf("%d files", len(m.files)))

	s += "\n"

	var l = lipgloss.NewStyle().Padding(0).Margin(0)

	var f string
	for i, file := range m.files {
		f += fmt.Sprintf("(%d) %s \n", i+1, file)
	}
	l.Render(f)

	s += f

	return s
}
