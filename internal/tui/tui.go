package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sectore/fit-sum-tui/internal/fit"
)

type Model struct {
	importPath   string
	files        []string
	errMsgs      []error
	selectedFile string
}

func InitialModel(path string) Model {
	return Model{
		importPath:   path,
		files:        nil,
		selectedFile: "",
	}
}

func (m Model) Init() tea.Cmd {
	return getFilesCmd(m.importPath)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		default:
			return m, nil

		}
	case filesMsg:
		m.files = msg

	case errMsg:
		m.errMsgs = append(m.errMsgs, msg)
	}
	return m, nil
}

func (m Model) View() string {

	s := fmt.Sprintf("path: %s\n", m.importPath)

	// headlien style
	var hs = lipgloss.NewStyle().
		Border(lipgloss.ThickBorder(), true).PaddingLeft(2).PaddingRight(2).MarginBottom(1)

	s += hs.Render(fmt.Sprintf("%d FIT files", len(m.files)))

	s += "\n"
	// errorStyle
	var es = lipgloss.NewStyle().Foreground(lipgloss.Color("#F25D94"))

	for _, err := range m.errMsgs {
		s += es.Render(fmt.Sprintf("%s\n", err))
	}
	// list style
	var ls = lipgloss.NewStyle().Padding(0).Margin(0)

	var f string
	for i, file := range m.files {
		f += fmt.Sprintf("(%d) %s \n", i+1, file)
	}
	ls.Render(f)

	s += f

	return s
}

type filesMsg []string

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

func getFilesCmd(path string) tea.Cmd {
	return func() tea.Msg {
		files, err := fit.GetFitFiles(path)
		if err != nil {
			return errMsg{err}
		}
		return filesMsg(files)

	}
}
