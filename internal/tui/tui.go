package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muktihari/fit/profile/filedef"
	"github.com/sectore/fit-sum-tui/internal/fit"
)

type Model struct {
	importPath   string
	files        []string
	activities   []*filedef.Activity
	errMsgs      []error
	selectedFile string
}

func InitialModel(path string) Model {
	return Model{
		importPath:   path,
		files:        nil,
		activities:   nil,
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
		return m, parseFilesCmd(msg)

	case activitiesMsg:
		m.activities = msg
		return m, nil

	case errMsg:
		m.errMsgs = append(m.errMsgs, msg)
		return m, nil
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
	s += ls.Render(f)

	s += fmt.Sprintf("no acts (%d) \n", len(m.activities))

	var a string
	for i, act := range m.activities {
		a += fmt.Sprintf("(%d) %s %s %s \n",
			i+1,
			fit.GetLocalTime(act).Format("2006-01-02 15:04"),
			fit.FormatTotalTime(act),
			fit.FormatTotalDistance(act))
	}
	s += ls.Render(a)

	return s
}

type filesMsg []string
type activitiesMsg []*filedef.Activity

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

func parseFilesCmd(files []string) tea.Cmd {
	return func() tea.Msg {
		activities, err := fit.ParseFiles(files)
		if err != nil {
			return errMsg{err}
		}
		return activitiesMsg(activities)

	}
}
