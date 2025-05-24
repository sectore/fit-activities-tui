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
	importPath string
	activities common.Activities
	errMsgs    []error
	spinner    spinner.Model
}

func InitialModel(path string) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return Model{
		importPath: path,
		activities: common.Activities{},
		spinner:    s,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, getFilesCmd(m.importPath))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "down":
			if !ActivitiesParsing(m.activities) {
				m.activities.Next()
			}
			return m, nil
		case "up":
			if !ActivitiesParsing(m.activities) {
				m.activities.Prev()
			}
			return m, nil
		case " ":
			if act, ok := m.activities.CurrentAct(); ok {
				act.Toggle()
			}
			return m, nil
		default:
			return m, nil
		}

	case getFilesResultMsg:
		activities := make([]common.Activity, len(msg))
		for i, path := range msg {
			activities[i] = common.Activity{
				Path: path,
				Data: asyncdata.NotAsked[error, common.ActivityData](),
			}
		}
		m.activities = common.NewActivities(activities)
		if act, ok := m.activities.CurrentAct(); ok {
			act.Data = common.ActivityLoading(nil)
			cmds = append(cmds, parseFileCmd(act))
		}

	case parseFileResultMsg:
		if !m.activities.IsLastIndex() {
			if act, ok := m.activities.Next(); ok {
				act.Data = common.ActivityLoading(nil)
				cmds = append(cmds, parseFileCmd(act))
			}
		} else {
			m.activities.FirstIndex()
		}

	case errMsg:
		m.errMsgs = append(m.errMsgs, msg)

	case spinner.TickMsg:
		s, cmd := m.spinner.Update(msg)
		m.spinner = s
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {

	s := fmt.Sprintf("path: %s\n", m.importPath)

	// headlien style
	var hs = lipgloss.NewStyle().
		Border(lipgloss.InnerHalfBlockBorder(), true).Padding(1).PaddingTop(0).PaddingBottom(0)

	loading := "  "
	if ActivitiesParsing(m.activities) {
		loading = m.spinner.View()
	}

	s += hs.Render(
		fmt.Sprintf("%s %d/%d FIT files (%d errors) i=%d",
			loading,
			ActivitiesParsed(m.activities)+ActivitiesFailures(m.activities),
			len(m.activities.All()),
			ActivitiesFailures(m.activities),
			m.activities.CurrentIndex(),
		),
	)

	// errorStyle
	var es = lipgloss.NewStyle().Foreground(lipgloss.Color("#F25D94"))

	for _, err := range m.errMsgs {
		s += es.Render(fmt.Sprintf("%s\n", err))
	}
	// list style
	var ls = lipgloss.NewStyle().MarginTop(1)
	var lss = ls.Background(lipgloss.Color("10"))

	for i, act := range m.activities.All() {
		if data, ok := asyncdata.GetSuccess(act.Data); ok {
			text := fmt.Sprintf("(%d) %s %s %s",
				i+1,
				data.LocalTime.Format("2006-01-02 15:04"),
				FormatTotalTime(data),
				FormatTotalDistance(data))
			if act.IsSelected() {
				s += lss.Render(text)
			} else {
				s += ls.Render(text)
			}
		}
	}

	return s
}

type getFilesResultMsg []string
type parseFileResultMsg struct{}

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

func getFilesCmd(path string) tea.Cmd {
	return func() tea.Msg {
		files, err := fit.GetFitFiles(path)
		if err != nil {
			return errMsg{err}
		}
		return getFilesResultMsg(files)

	}
}

func parseFileCmd(act *common.Activity) tea.Cmd {
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
			resultCh <- parseFileResultMsg{}
			close(resultCh)
		}()
		// return result msg
		return <-resultCh
	}
}
