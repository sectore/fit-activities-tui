package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sectore/fit-sum-tui/internal/asyncdata"
	"github.com/sectore/fit-sum-tui/internal/common"
	"github.com/sectore/fit-sum-tui/internal/fit"
)

type Model struct {
	importPath       string
	activities       []common.Activity
	errMsgs          []error
	selectedFile     string
	currentFileIndex int
	spinner          spinner.Model
}

func InitialModel(path string) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return Model{
		importPath:       path,
		activities:       nil,
		selectedFile:     "",
		currentFileIndex: 0,
		spinner:          s,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(getFilesCmd(m.importPath), m.spinner.Tick)
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
		var activities []common.Activity
		for _, path := range msg {
			activities = append(activities, common.Activity{
				Path: path,
				Data: asyncdata.NotAsked[error, common.ActivityData](),
			})
		}
		m.activities = activities
		m.activities[0].Data = common.ActivityLoading(nil)
		return m, parseFileCmd(m.activities[0])

	case parseFileResultMsg:
		m.activities[m.currentFileIndex] = msg.activity
		if m.currentFileIndex < len(m.activities)-1 {
			m.currentFileIndex++
			m.activities[m.currentFileIndex].Data = common.ActivityLoading(nil)
			return m, parseFileCmd(m.activities[m.currentFileIndex])
		}
		return m, nil

	case errMsg:
		m.errMsgs = append(m.errMsgs, msg)
		return m, nil
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m Model) View() string {

	s := fmt.Sprintf("path: %s\n", m.importPath)

	// headlien style
	var hs = lipgloss.NewStyle().
		Border(lipgloss.ThickBorder(), true).PaddingLeft(2).PaddingRight(2).MarginBottom(1)

	loading := "  "
	if ActivitiesAreLoading(m.activities) {
		loading = m.spinner.View()
	}

	s += hs.Render(
		fmt.Sprintf("%s parse %d/%d FIT files (%d errors)",
			loading,
			ActivitiesSuccess(m.activities)+ActivitiesFailures(m.activities),
			len(m.activities),
			ActivitiesFailures(m.activities),
		),
	)

	s += "\n"

	// errorStyle
	var es = lipgloss.NewStyle().Foreground(lipgloss.Color("#F25D94"))

	for _, err := range m.errMsgs {
		s += es.Render(fmt.Sprintf("%s\n", err))
	}
	// list style
	var ls = lipgloss.NewStyle().Padding(0).Margin(0)

	var a string
	for i, act := range m.activities {
		data, ok := asyncdata.GetSuccess(act.Data)
		if ok {

			a += fmt.Sprintf("(%d) %s %s %s \n",
				i+1,
				data.LocalTime.Format("2006-01-02 15:04"),
				FormatTotalTime(data),
				FormatTotalDistance(data))
		}

	}
	s += ls.Render(a)

	return s
}

type filesMsg []string
type parseFileResultMsg struct{ activity common.Activity }

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

func parseFileCmd(act common.Activity) tea.Cmd {
	return func() tea.Msg {
		// channel to send result msg
		resultCh := make(chan tea.Msg, 1)
		// goroutine to do the parsing
		go func() {
			data, err := fit.ParseFile(act.Path)
			if err != nil {
				act.Data = common.ActivityFailure(err)
			} else {
				act.Data = common.ActivitySuccess(*data)
			}
			// FIXME: for debugging only
			time.Sleep(100 * time.Millisecond)
			resultCh <- parseFileResultMsg{act}
			close(resultCh)
		}()
		// return result msg
		return <-resultCh
	}
}
