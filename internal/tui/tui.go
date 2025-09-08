package tui

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/sectore/fit-activities-tui/internal/asyncdata"
	"github.com/sectore/fit-activities-tui/internal/common"
	"github.com/sectore/fit-activities-tui/internal/fit"
)

type ActsSort int

const (
	NoSort ActsSort = iota
	TimeAsc
	TimeDesc
	DistanceAsc
	DistanceDesc
)

type Model struct {
	importPath  string
	importIndex int
	activities  common.Activities
	errMsgs     []error
	spinner     spinner.Model
	list        list.Model
	width       int
	height      int
	showMenu    bool
	actsSort    ActsSort
}

const (
	arrowTop       = "↑"
	arrowDown      = "↓"
	BulletPointBig = "●"
	BulletPoint    = "∙"
)

var (
	contentStyle      = lipgloss.NewStyle().Padding(1)
	leftContentStyle  = lipgloss.NewStyle().Width(30)
	rightContentStyle = lipgloss.NewStyle().Padding(0, 2, 2, 0)
	footerStyle       = lipgloss.NewStyle().Padding(0, 1)
	errorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#F25D94"))
	noColor           = lipgloss.NoColor{}
	emptyStyle        = lipgloss.NewStyle()
	br                = lipgloss.NewStyle().SetString("\n").String()
)

func InitialModel(path string) Model {

	s := spinner.New()
	s.Spinner = spinner.MiniDot
	// Note: We do need to pass `Spinner` down to the `ListDelegate` of the list
	// to make sure `spinner.Tick` is fired once. Currently in `Init`.
	delegate := NewListDelegate(&s)

	list := list.New([]list.Item{}, &delegate, 20, 0)
	list.Title = ""

	// styles for prompt needs to be passed to `FilterInput`
	fi := list.FilterInput
	fi.Prompt = "/"
	fi.PromptStyle = emptyStyle
	fi.Cursor.Style = emptyStyle
	list.FilterInput = fi

	// styles for paginator needs to be passed to `Paginator`
	p := list.Paginator
	p.ActiveDot = lipgloss.NewStyle().SetString(BulletPointBig).String()
	p.InactiveDot = lipgloss.NewStyle().SetString(BulletPoint).String()
	list.Paginator = p

	ls := list.Styles
	ls.Title = emptyStyle
	ls.DividerDot = list.Styles.DividerDot.Foreground(noColor)
	ls.StatusBar = list.Styles.StatusBar.Foreground(noColor)
	ls.StatusEmpty = emptyStyle
	ls.StatusBarActiveFilter = emptyStyle
	ls.StatusBarFilterCount = emptyStyle
	ls.NoItems = emptyStyle

	list.Styles = ls

	list.SetSpinner(spinner.MiniDot)
	list.SetShowHelp(false)
	list.SetShowTitle(true)
	list.SetShowStatusBar(false)

	return Model{
		importPath:  path,
		importIndex: 0,
		activities:  common.Activities{},
		spinner:     s,
		list:        list,
		width:       0,
		height:      0,
		showMenu:    false,
		actsSort:    NoSort,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, getFilesCmd(m.importPath))
}

func (m *Model) sortActs() tea.Cmd {
	acts := ListItemsToActivities(m.list.Items())

	switch m.actsSort {
	case DistanceAsc:
		common.SortBy(common.SortByDistance).Sort(acts)
	case DistanceDesc:
		common.SortBy(common.SortByDistance).Reverse(acts)
	case TimeAsc:
		common.SortBy(common.SortByTime).Sort(acts)
	case TimeDesc:
		common.SortBy(common.SortByTime).Reverse(acts)
	default:
		// do nothing
	}

	items := ActivitiesToListItems(acts)
	// Note: `SetItems` resets the filter internally.
	// That's remember filter text BEFORE ...
	filterText := m.list.FilterInput.Value()
	// ... updating ALL items ...
	cmd := m.list.SetItems(items)
	// ... and set filter text again to set filter internally.
	if filterText != "" {
		m.list.SetFilterText(filterText)
	}

	return cmd
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case tea.KeyMsg:
		log.Printf("key %s", msg.String())
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
				cmd := m.list.SetItems([]list.Item{})

				cmds = append(cmds, cmd, getFilesCmd(m.importPath))

			}
		case "m":
			if !m.list.SettingFilter() {
				m.showMenu = !m.showMenu
			}
		case "ctrl+d":
			if !ActivitiesParsing(m.activities) {
				if m.actsSort != DistanceDesc {
					m.actsSort = DistanceDesc
				} else {
					m.actsSort = DistanceAsc
				}
				cmd := m.sortActs()
				cmds = append(cmds, cmd)
			}
		case "ctrl+t":
			if !ActivitiesParsing(m.activities) {
				if m.actsSort != TimeDesc {
					m.actsSort = TimeDesc
				} else {
					m.actsSort = TimeAsc
				}
				cmd := m.sortActs()
				cmds = append(cmds, cmd)
			}
		case "q":
			if !m.list.SettingFilter() {
				return m, tea.Quit
			}
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
		items := ActivitiesToListItems(activities)
		cmd := m.list.SetItems(items)
		// parse first Activity
		firstAct := m.activities[0]
		firstAct.Data = asyncdata.NewLoading[error, common.ActivityData](nil)
		cmds = append(cmds, cmd, parseFileCmd(firstAct))

	case parseFileResultMsg:
		i := m.importIndex
		m.activities[i] = msg.Activity
		cmd := m.list.SetItem(i, msg.Activity)
		cmds = append(cmds, cmd)

		if i < len(m.activities)-1 {
			m.importIndex++
			act := m.activities[m.importIndex]
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

func (m Model) RightContentView() string {

	var sumView string
	visibleItems := ListItemsToActivities(m.list.VisibleItems())
	noVisibleActs := len(visibleItems)
	if noVisibleActs > 0 {
		label := "activities"
		if noVisibleActs <= 1 {
			label = "activity"
		}
		labelNoActs := fmt.Sprintf("%d", noVisibleActs)
		label = fmt.Sprintf("%s %s", labelNoActs, label)
		sumView += lipgloss.NewStyle().
			PaddingRight(4).
			Bold(true).
			Border(lipgloss.ASCIIBorder(), false, false, true, false).
			Render(label)

		sumView += br

		sumRows := [][]string{
			{"total time", ActivitiesTotalDuration(visibleItems).Format()},
			{"total distance", ActivitiesTotalDistances(visibleItems).Format()},
		}
		sumTable := table.New().
			Rows(sumRows...).
			Border(lipgloss.Border{}).
			StyleFunc(func(row, col int) lipgloss.Style {
				switch col {
				case 0:
					return lipgloss.NewStyle().PaddingRight(2).Bold(true)
				default:
					return emptyStyle
				}
			})
		sumView += fmt.Sprintf("%s", sumTable)
	} else {
		sumView += "No activity found."
	}

	var detailsView string
	item := m.list.SelectedItem()
	if item != nil && !m.list.SettingFilter() {
		detailsView += lipgloss.NewStyle().
			Bold(true).
			PaddingRight(4).
			Border(lipgloss.ASCIIBorder(), false, false, true, false).
			MarginBottom(1).
			Render(fmt.Sprintf(`#%d activity`, m.list.Index()+1))

		rows := [][]string{
			{"date", "..."},
			{"distance", "..."},
			{"time", "..."},
			{"speed", "..."},
			{"elevation", "..."},
			{"temperature", "..."},
			{"gps accuracy", "..."},
			{"sessions", "..."},
			{"records", "..."},
		}
		var col = lipgloss.NewStyle().PaddingRight(3).Render
		if act, ok := item.(common.Activity); ok {
			if act, ok := asyncdata.Success(act.Data); ok {
				// date
				rows[0][1] = act.StartTime.Format()
				// distance
				rows[1][1] = act.TotalDistance.Format()
				// time
				rows[2][1] = fmt.Sprintf(`active %s pause %s Σ %s`,
					col(act.Duration.Active.Format()),
					col(act.Duration.Pause.Format()),
					act.Duration.Total.Format(),
				)
				// speed
				rows[3][1] = fmt.Sprintf(`⌀ %s %s %s`,
					col(act.Speed.Avg.Format()),
					arrowTop,
					act.Speed.Max.Format())

				// Elevation
				rows[4][1] = fmt.Sprintf(`%s %s %s %s`,
					arrowTop,
					col(act.Elevation.Ascents.Format()),
					arrowDown,
					act.Elevation.Descents.Format(),
				)
				// temperature
				rows[5][1] = fmt.Sprintf(`⌀ %s %s %s %s %s`,
					col(act.Temperature.Avg.Format()),
					arrowDown,
					col(act.Temperature.Min.Format()),
					arrowTop,
					act.Temperature.Max.Format(),
				)
				// gps
				rows[6][1] = fmt.Sprintf(`⌀ %s %s %s %s %s`,
					col(act.GpsAccuracy.Avg.Format()),
					arrowDown,
					col(act.GpsAccuracy.Min.Format()),
					arrowTop,
					act.GpsAccuracy.Max.Format(),
				)
				// no. sessions
				rows[7][1] = fmt.Sprintf(`%d`, act.NoSessions)
				// no. records
				rows[8][1] = fmt.Sprintf(`%d`, act.NoRecords)
			}
			rows = append(rows,
				[]string{"file", filepath.Base(act.Path)},
			)
			table := table.New().
				Rows(rows...).
				Border(lipgloss.Border{}).
				StyleFunc(func(row, col int) lipgloss.Style {
					switch {
					case col == 0:
						return lipgloss.NewStyle().PaddingRight(2).Bold(true)
					case row == 2 || row == 6:
						return lipgloss.NewStyle().MarginBottom(1)
					default:
						return lipgloss.NewStyle().PaddingRight(1)
					}
				})
			detailsView += fmt.Sprintf("%s", table)
		}

	}

	return lipgloss.JoinVertical(lipgloss.Left,
		sumView,
		lipgloss.NewStyle().
			MarginTop(2).
			Render(detailsView),
	)
}

func (m Model) LeftContentView() string {

	noVisibleActs := len(m.list.VisibleItems())
	noActs := len(m.list.Items())

	if m.list.IsFiltered() && noVisibleActs != noActs {
		labelNoActs := fmt.Sprintf("%d of %d", noVisibleActs, len(m.list.Items()))
		m.list.Title = lipgloss.NewStyle().Bold(false).Italic(true).Render(labelNoActs)
	} else {
		m.list.Title = lipgloss.NewStyle().Bold(true).Render("All")
	}

	return m.list.View()
}

func (m Model) footerView() string {
	symbol := arrowTop
	if m.showMenu {
		symbol = arrowDown
	}
	menu := fmt.Sprintf("[m]enu %s", symbol)
	line := strings.Repeat("─", max(0, m.width-len(menu)-1))
	view := fmt.Sprintf("%s %s", menu, line)
	if m.showMenu {
		filterCol2 := "[/]start filter"
		if m.list.SettingFilter() {
			if len(m.list.FilterValue()) > 0 {
				filterCol2 = "[ENTER]apply filter"
			} else {
				filterCol2 = "[ESC]cancel filter"
			}
		}
		filterCol3 := "[^d]sort by distance"
		filterCol4 := "[^t]sort by start time"
		if m.list.IsFiltered() {
			filterCol3 = "[ESC]clear filter"
			filterCol4 = ""
		}
		if m.list.SettingFilter() {
			filterCol3 = ""
			filterCol4 = ""
			if len(m.list.FilterValue()) > 0 {
				filterCol3 = "[ESC]skip filter"
			}
		}
		rows := [][]string{
			{"actions", "[r]eload", "[q]uit"},
			{"list", "[" + arrowTop + "]up", "[" + arrowDown + "]down", "[g]first", "[G]last", "[←/→]switch pages"},
			{"filter/sort", filterCol2, filterCol3, filterCol4},
		}
		table := table.New().
			Rows(rows...).
			Border(lipgloss.Border{}).
			StyleFunc(func(row, col int) lipgloss.Style {
				switch {
				case col == 0:
					return lipgloss.NewStyle().PaddingRight(5).Bold(true)
				default:
					return lipgloss.NewStyle().PaddingRight(2)
				}
			})
		view += fmt.Sprintf("%s", table)
	}

	return view
}

func (m Model) View() string {

	footer := m.footerView()
	footerH := lipgloss.Height(footer)
	contentHeight := m.height - footerH - 2 // add padding of content
	m.list.SetHeight(contentHeight)
	content := lipgloss.JoinHorizontal(lipgloss.Top,
		leftContentStyle.Render(m.LeftContentView()),
		rightContentStyle.
			MaxHeight(contentHeight).
			Render(m.RightContentView()),
	)
	view := lipgloss.JoinVertical(lipgloss.Position(lipgloss.Left),
		contentStyle.Render(content),
		footerStyle.Render(footer))

	return view
}

type (
	getFilesResultMsg  []string
	parseFileResultMsg struct{ common.Activity }
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

func parseFileCmd(act common.Activity) tea.Cmd {
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
			// time.Sleep(50 * time.Millisecond)
			resultCh <- parseFileResultMsg{act}
			close(resultCh)
		}()
		// return result msg
		return <-resultCh
	}
}
