package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
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
	list       list.Model
}

func InitialModel(path string) Model {

	s := spinner.New()
	s.Spinner = spinner.Dot

	list := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)

	return Model{
		importPath: path,
		activities: common.Activities{},
		spinner:    s,
		list:       list,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, getFilesCmd(m.importPath))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-6)

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		}

	case getFilesResultMsg:
		activities := make([]common.Activity, len(msg))
		for i, path := range msg {
			activities[i] = common.Activity{
				Path: path,
				Data: asyncdata.NewNotAsked[error, common.ActivityData](),
			}
		}
		m.activities = common.NewActivities(activities)

		items := make([]list.Item, len(msg))
		m.list.SetItems(items)

		if act, ok := m.activities.CurrentAct(); ok {
			act.Data = asyncdata.NewLoading[error, common.ActivityData](nil)
			cmds = append(cmds, parseFileCmd(act))
		}

	case parseFileResultMsg:
		current := m.activities.CurrentIndex()
		if cAct, ok := m.activities.CurrentAct(); ok {
			m.list.SetItem(int(current), cAct)
		}
		if !m.activities.IsLastIndex() {

			if act, ok := m.activities.Next(); ok {
				act.Data = asyncdata.NewLoading[error, common.ActivityData](nil)
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

	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

var (
	appStyle          = lipgloss.NewStyle().Padding(1, 2)
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
)

func (m Model) View() string {

	s := m.list.View()

	s += "\n"

	// headline style
	var hs = lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(3).
		MarginLeft(2).
		Border(lipgloss.NormalBorder(), true, false)

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
	s += "\n"
	s += lipgloss.NewStyle().
		PaddingLeft(2).
		Render(fmt.Sprintf("%s", m.importPath))
	s += "\n"

	// errorStyle
	var es = lipgloss.NewStyle().Foreground(lipgloss.Color("#F25D94"))

	for _, err := range m.errMsgs {
		s += es.Render(fmt.Sprintf("%s\n", err))
	}

	s += "\n"

	return appStyle.Render(s)
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
				act.Data = asyncdata.NewFailure[error, common.ActivityData](err)
			} else {
				act.Data = asyncdata.NewSuccess[error, common.ActivityData](*data)
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
