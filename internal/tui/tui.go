package tui

import (
	"fmt"
	"io"
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

type listDelegate struct {
	DefaultDelegate list.DefaultDelegate
	Spinner         spinner.Model
}

func (d listDelegate) Height() int  { return 2 }
func (d listDelegate) Spacing() int { return 1 }

func NewListDelegate(spinner *spinner.Model) listDelegate {

	s := list.NewDefaultItemStyles()
	s.NormalTitle = lipgloss.NewStyle().
		Padding(0, 0, 0, 2) //nolint:mnd
	s.DimmedTitle = s.NormalTitle
	s.NormalDesc = s.NormalTitle
	s.DimmedDesc = s.NormalDesc
	selectedStyle := lipgloss.NewStyle().
		Border(lipgloss.OuterHalfBlockBorder(), false, false, false, true).
		Padding(0, 0, 0, 1)
	s.SelectedTitle = selectedStyle.
		Bold(true)
	s.SelectedDesc = selectedStyle
	s.FilterMatch = lipgloss.NewStyle().Bold(true)

	d := list.NewDefaultDelegate()

	d.Styles = s

	cd := listDelegate{DefaultDelegate: d, Spinner: *spinner}
	return cd
}

func (d *listDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	// `ActivityAD` is our custom `Item`
	if act, ok := item.(common.Activity); ok {
		_, _, loading := asyncdata.Loading(act.Data)
		notAsked := asyncdata.NotAsked(act.Data)
		if loading || notAsked {
			d.Spinner.Style = lipgloss.NewStyle().MarginBottom(1).MarginLeft(2)
			fmt.Fprintf(w, "%s", d.Spinner.View())
			return
		}
	}
	// TODO: render `Failure`

	// use default render
	d.DefaultDelegate.Render(w, m, index, item)
}

// Delegate `Update` to have still an animated spinner for each item
func (d *listDelegate) Update(msg tea.Msg, _ *list.Model) tea.Cmd {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		s, cmd := d.Spinner.Update(msg)
		d.Spinner = s
		return cmd
	}
	return nil
}

const (
	pageActiveBullet   = "●"
	pageInactiveBullet = "∙"
)

func InitialModel(path string) Model {

	s := spinner.New()
	s.Spinner = spinner.MiniDot
	// Note: We do need to pass `Spinner` down to the `ItemDelegate` of the list
	// to make sure `spinner.Tick` is fired once. Currently in `Init`.
	delegate := NewListDelegate(&s)
	list := list.New([]list.Item{}, &delegate, 20, 0)
	list.Title = "Activities"

	// noForeground := lipgloss.NewStyle().Foreground(lipgloss.NoColor{})
	noStyle := lipgloss.NewStyle()

	// styles for prompt needs to be passed to `FilterInput`
	fi := list.FilterInput
	fi.Prompt = "/"
	fi.PromptStyle = noStyle
	fi.Cursor.Style = noStyle
	list.FilterInput = fi

	// styles for paginator needs to be passed to `Paginator`
	p := list.Paginator
	p.ActiveDot = lipgloss.NewStyle().SetString(pageActiveBullet).String()
	p.InactiveDot = lipgloss.NewStyle().SetString(pageInactiveBullet).String()
	list.Paginator = p

	ls := list.Styles
	ls.Title = lipgloss.NewStyle().Bold(true)
	ls.DividerDot = list.Styles.DividerDot.Foreground(lipgloss.NoColor{})
	ls.StatusBar = list.Styles.StatusBar.Foreground(lipgloss.NoColor{})
	ls.StatusEmpty = noStyle
	ls.StatusBarActiveFilter = noStyle
	ls.StatusBarFilterCount = noStyle
	ls.NoItems = noStyle

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

var (
	appStyle          = lipgloss.NewStyle().Padding(1, 2)
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
)

type Content struct {
	common.Activities
}

func renderContent(m Model) string {
	var content string = ""
	item := m.list.SelectedItem()
	if item != nil {
		// Note: Item is a Pointer here !!!
		if act, ok := item.(*common.Activity); ok {
			if act, ok := asyncdata.Success[error, common.ActivityData](act.Data); ok {
				content = fmt.Sprintf("total time \n%s\n\n", act.FormatTotalTime())
			}
			content += fmt.Sprintf("file\n%s", filepath.Base(act.Path))
		}

	}
	return content
}

func (m Model) View() string {

	s := lipgloss.JoinHorizontal(lipgloss.Top, m.list.View(), lipgloss.NewStyle().Padding(2).MarginTop(2).Render(
		fmt.Sprintf("%s", renderContent(m))),
	)

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
			len(m.activities),
			ActivitiesFailures(m.activities),
			m.list.Index(),
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
