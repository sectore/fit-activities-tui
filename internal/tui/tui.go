package tui

import (
	"fmt"
	"path/filepath"
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
	importPath  string
	importIndex int
	activities  common.Activities
	errMsgs     []error
	spinner     spinner.Model
	list        list.Model
}

const (
	pageActiveBullet   = "●"
	pageInactiveBullet = "∙"
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)
	// headline style
	hStyle = lipgloss.NewStyle().
		PaddingLeft(1).
		PaddingRight(3).
		MarginLeft(2).
		Border(lipgloss.NormalBorder(), true, false)
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F25D94"))
	noColor    = lipgloss.NoColor{}
	emptyStyle = lipgloss.NewStyle()
	br         = lipgloss.NewStyle().SetString("\n").String()
)

func InitialModel(path string) Model {

	s := spinner.New()
	s.Spinner = spinner.MiniDot
	// Note: We do need to pass `Spinner` down to the `ListDelegate` of the list
	// to make sure `spinner.Tick` is fired once. Currently in `Init`.
	delegate := NewListDelegate(&s)

	list := list.New([]list.Item{}, &delegate, 20, 0)
	list.Title = "Activities"

	// noForeground := lipgloss.NewStyle().Foreground(lipgloss.NoColor{})

	// styles for prompt needs to be passed to `FilterInput`
	fi := list.FilterInput
	fi.Prompt = "/"
	fi.PromptStyle = emptyStyle
	fi.Cursor.Style = emptyStyle
	list.FilterInput = fi

	// styles for paginator needs to be passed to `Paginator`
	p := list.Paginator
	p.ActiveDot = lipgloss.NewStyle().SetString(pageActiveBullet).String()
	p.InactiveDot = lipgloss.NewStyle().SetString(pageInactiveBullet).String()
	list.Paginator = p

	ls := list.Styles
	ls.Title = lipgloss.NewStyle().Bold(true)
	ls.DividerDot = list.Styles.DividerDot.Foreground(noColor)
	ls.StatusBar = list.Styles.StatusBar.Foreground(noColor)
	ls.StatusEmpty = emptyStyle
	ls.StatusBarActiveFilter = emptyStyle
	ls.StatusBarFilterCount = emptyStyle
	ls.NoItems = emptyStyle

	list.Styles = ls

	list.SetSpinner(spinner.MiniDot)
	list.SetStatusBarItemName("activity", "activities")
	list.SetShowHelp(false)

	return Model{
		importPath:  path,
		importIndex: 0,
		activities:  common.Activities{},
		spinner:     s,
		list:        list,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, getFilesCmd(m.importPath))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		_, v := appStyle.GetFrameSize()
		m.list.SetHeight(msg.Height - v - 6)

	case tea.KeyMsg:
		switch msg.String() {
		// reload data
		case "r":
			// ignore if an user typing a filter ...
			if !m.list.SettingFilter() {
				// reset activities
				m.activities = common.Activities{}
				// reset import index
				m.importIndex = 0
				// reset list
				m.list.ResetSelected()
				m.list.ResetFilter()
				m.list.SetItems([]list.Item{})

				cmds = append(cmds, getFilesCmd(m.importPath))

			}
		case "q":
			return m, tea.Quit
		}

	case getFilesResultMsg:
		// list of `NotAsked` activities
		activities := make([]common.Activity, len(msg))
		for i, path := range msg {
			activities[i] = common.Activity{
				Path: path,
				Data: asyncdata.NewNotAsked[error, common.ActivityData](),
			}
		}
		m.activities = activities
		// transform activities to be `list.Item`
		items := make([]list.Item, len(msg))
		for i, act := range activities {
			items[i] = act
		}
		m.list.SetItems(items)
		// parse first Activity
		firstAct := m.activities[0]
		firstAct.Data = asyncdata.NewLoading[error, common.ActivityData](nil)
		cmds = append(cmds, parseFileCmd(&firstAct))

	case parseFileResultMsg:
		i := m.importIndex
		m.activities[i] = *msg.Activity
		m.list.SetItem(i, msg.Activity)

		if i < len(m.activities)-1 {
			m.importIndex++
			act := &m.activities[m.importIndex]
			act.Data = asyncdata.NewLoading[error, common.ActivityData](nil)
			cmds = append(cmds, parseFileCmd(act))

		}

	case errMsg:
		m.errMsgs = append(m.errMsgs, msg)

	case spinner.TickMsg:
		s, cmd := m.spinner.Update(msg)
		m.spinner = s
		cmds = append(cmds, cmd)
	}

	if ActivitiesParsing(m.activities) {
		cmd := m.list.StartSpinner()
		cmds = append(cmds, cmd)
	} else {
		m.list.StopSpinner()
	}

	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) contentView() string {
	var view string
	item := m.list.SelectedItem()
	if item != nil && !m.list.SettingFilter() {
		// Note: Item is a Pointer here !!!
		if act, ok := item.(*common.Activity); ok {
			if act, ok := asyncdata.Success[error, common.ActivityData](act.Data); ok {
				view += "total time %s"
				view += br
				view += act.FormatTotalTime()
				view += br
				view += br
			}
			view += "file"
			view += br
			view += filepath.Base(act.Path)
		}

	}
	return view
}

func (m Model) View() string {

	view := lipgloss.JoinHorizontal(lipgloss.Top, m.list.View(), lipgloss.NewStyle().Padding(2).MarginTop(2).Render(
		m.contentView()),
	)

	view += br

	loading := "  "
	if ActivitiesParsing(m.activities) {
		loading = m.spinner.View()
	}

	view += hStyle.Render(
		fmt.Sprintf("%s %d/%d FIT files (%d errors) i=%d",
			loading,
			ActivitiesParsed(m.activities)+ActivitiesFailures(m.activities),
			len(m.activities),
			ActivitiesFailures(m.activities),
			m.list.Index(),
		),
	)
	view += br
	view += lipgloss.NewStyle().
		PaddingLeft(2).
		Render(m.importPath)
	view += br

	for _, err := range m.errMsgs {
		view += errorStyle.Render(fmt.Sprintf("%s", err))
		view += br
	}

	return appStyle.Render(view)
}

type (
	getFilesResultMsg  []string
	parseFileResultMsg struct{ *common.Activity }
	errMsg             struct{ err error }
)

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
			time.Sleep(50 * time.Millisecond)
			resultCh <- parseFileResultMsg{act}
			close(resultCh)
		}()
		// return result msg
		return <-resultCh
	}
}
