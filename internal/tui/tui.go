package tui

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

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
	TimeAsc ActsSort = iota
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
	// live data
	showLiveData       bool
	playLiveData       bool
	liveDataSpeed      uint
	liveDataLastUpdate time.Time
}

const (
	FPS         = 60
	FPSDuration = time.Second / FPS

	LiveDataSpeedBoost = 60 * 5 // about 5min at 1RPS
	LiveDataMaxSpeed   = 10

	arrowTop       = "↑"
	arrowDown      = "↓"
	BulletPointBig = "●"
	BulletPoint    = "∙"
	BarEmpty       = "░"
	BarEmptyHalf   = "▒"
	BarFullHalf    = "▓"
	BarFull        = "█"
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

	l := list.New([]list.Item{}, &delegate, 20, 0)
	// Keep `Title` empty for now
	// It will be set (incl. styles) in `LeftContentView`
	l.Title = ""

	// Unbind `Page` keys (they don't work for any reason)
	// Now we can toggle `live data` using "l" key without to get in conflict anymore
	keyMap := list.DefaultKeyMap()
	keyMap.NextPage.Unbind()
	keyMap.PrevPage.Unbind()
	l.KeyMap = keyMap

	// styles for prompt needs to be passed to `FilterInput`
	lfi := l.FilterInput
	lfi.Prompt = "/"
	lfi.PromptStyle = emptyStyle
	lfi.Cursor.Style = emptyStyle
	l.FilterInput = lfi

	// styles for paginator needs to be passed to `Paginator`
	lp := l.Paginator
	lp.ActiveDot = lipgloss.NewStyle().SetString(BulletPointBig).String()
	lp.InactiveDot = lipgloss.NewStyle().SetString(BulletPoint).String()
	l.Paginator = lp

	ls := l.Styles
	ls.Title = emptyStyle
	ls.DividerDot = l.Styles.DividerDot.Foreground(noColor)
	ls.StatusBar = l.Styles.StatusBar.Foreground(noColor)
	ls.StatusEmpty = emptyStyle
	ls.StatusBarActiveFilter = emptyStyle
	ls.StatusBarFilterCount = emptyStyle
	ls.NoItems = emptyStyle

	l.Styles = ls

	l.SetSpinner(spinner.MiniDot)
	l.SetShowHelp(false)
	l.SetShowTitle(true)
	l.SetShowStatusBar(false)

	return Model{
		importPath:         path,
		importIndex:        0,
		activities:         common.Activities{},
		spinner:            s,
		list:               l,
		width:              0,
		height:             0,
		showMenu:           false,
		actsSort:           TimeDesc,
		showLiveData:       false,
		playLiveData:       false,
		liveDataSpeed:      1,
		liveDataLastUpdate: time.Now(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, getFilesCmd(m.importPath), tick())
}

func (m *Model) sortActs() tea.Cmd {

	items := SortItems(m.list.Items(), m.actsSort)

	// Note: `SetItems` resets the filter internally.
	// That's why we need to remember filter text BEFORE ...
	filterText := m.list.FilterInput.Value()
	// ... updating ALL items ...
	cmd := m.list.SetItems(items)
	// ... to set filter text again, which sets filter internally.
	if filterText != "" {
		m.list.SetFilterText(filterText)
	}

	return cmd
}

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(FPSDuration, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case tea.KeyMsg:
		log.Printf("key %s", msg.String())
		switch msg.String() {

		case "r":
			// reset `selectedRecordIndex` of SELECTED item
			if !m.list.SettingFilter() {
				item := m.list.SelectedItem()
				if act, ok := item.(*common.Activity); ok {
					act.ResetRecordIndex()
				}
			}
		case "ctrl+r":
			// reset `selectedRecordIndex` of ALL items
			if !m.list.SettingFilter() {
				for _, item := range m.list.Items() {
					if act, ok := item.(*common.Activity); ok {
						act.ResetRecordIndex()
					}
				}
			}
		// reload data
		case "alt+ctrl+r":
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
		case "l":
			m.showLiveData = !m.showLiveData
		case " ":
			if m.showLiveData {
				m.playLiveData = !m.playLiveData
				if m.playLiveData {
					m.liveDataLastUpdate = time.Now()
				}
			}
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			if m.showLiveData && !m.list.SettingFilter() {
				// Convert ASCII values to integer:
				// 1. msg.String()[0]` gets the first character of the key press as a byte
				// 2. `- '0'` subtracts the ASCII value of '0' (48) from it
				// 3. `int()` converts the result to an integer
				// Examples:
				// - `'1' - '0'` = 49 - 48 = 1
				// - `'5' - '0'` = 53 - 48 = 5
				// - `'9' - '0'` = 57 - 48 = 9
				speed := uint(msg.String()[0] - '0')
				if speed == 0 {
					speed = LiveDataMaxSpeed
				}
				m.liveDataSpeed = speed
			}
		case "left":
			if m.showLiveData && !m.list.SettingFilter() {
				if m.playLiveData {
					// playing: decrease speed
					if m.liveDataSpeed > 1 {
						m.liveDataSpeed--
					}
				} else {
					// pause: skip back to previous `Record`
					item := m.list.SelectedItem()
					if act, ok := item.(*common.Activity); ok {
						act.CountRecordIndex(-1)
					}
				}
			}
		case "ctrl+left":
			if m.showLiveData &&
				!m.list.SettingFilter() {
				// pause only (ignore playing):
				// skip to a previous `Record` (boosted)
				if !m.playLiveData {
					item := m.list.SelectedItem()
					if act, ok := item.(*common.Activity); ok {
						back := -LiveDataSpeedBoost
						act.CountRecordIndex(back)
					}
				}
			}
		case "right":
			if m.showLiveData && !m.list.SettingFilter() {
				// playing: increase speed
				if m.playLiveData {
					if m.liveDataSpeed < 10 {
						m.liveDataSpeed++
					}
				} else {
					// pause: skip to next `Record`
					item := m.list.SelectedItem()
					if act, ok := item.(*common.Activity); ok {
						act.CountRecordIndex(1)
					}
				}
			}

		case "ctrl+right":
			if m.showLiveData &&
				!m.list.SettingFilter() {
				// playing: increase speed (boosted)
				if m.playLiveData {
					if m.liveDataSpeed <= LiveDataMaxSpeed {
						m.liveDataSpeed += LiveDataSpeedBoost
					}
				} else {
					// pause: skip to a next `Record` (boosted)
					item := m.list.SelectedItem()
					if act, ok := item.(*common.Activity); ok {
						next := LiveDataSpeedBoost
						act.CountRecordIndex(next)
					}
				}
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
		activities := make([]*common.Activity, len(msg))
		for i, path := range msg {
			activities[i] = &common.Activity{
				Path: path,
				Data: asyncdata.NewNotAsked[error, common.ActivityData](),
			}
		}
		m.activities = activities
		// parse first Activity
		firstAct := m.activities[0]
		firstAct.Data = asyncdata.NewLoading[error, common.ActivityData](nil)
		cmds = append(cmds, parseFileCmd(firstAct))

	case parseFileResultMsg:
		i := m.importIndex
		*m.activities[i] = *msg.Activity
		// add new item to current items
		items := append(m.list.Items(), m.activities[i])
		// sort list
		items = SortItems(items, m.actsSort)
		// set sorted items to list
		cmd := m.list.SetItems(items)
		cmds = append(cmds, cmd)

		if i < len(m.activities)-1 {
			m.importIndex++
			act := m.activities[m.importIndex]
			act.Data = asyncdata.NewLoading[error, common.ActivityData](nil)
			cmds = append(cmds, parseFileCmd(act))
		}

	case errMsg:
		m.errMsgs = append(m.errMsgs, msg)

	case tickMsg:
		now := time.Now()

		// live data playback
		if m.playLiveData {
			item := m.list.SelectedItem()
			if act, ok := item.(*common.Activity); ok {
				elapsed := now.Sub(m.liveDataLastUpdate).Milliseconds()
				if elapsed > 0 {
					// transform RPS -> RPmS (records per millisecond)
					rpms := act.RPS() / 1000.0
					// For "smooth" updates:
					// Calculate how many records to advance based on elapsed time and speed
					recordsToAdvance := int(float64(elapsed) * rpms * float64(m.liveDataSpeed))

					if recordsToAdvance > 0 {
						_ = act.CountRecordIndex(recordsToAdvance)
						m.liveDataLastUpdate = now
					}
				}
			}
		}

		// Reset speed boost an user might done before
		// so it will be for one "tick" available only
		if m.liveDataSpeed > LiveDataMaxSpeed {
			m.liveDataSpeed = m.liveDataSpeed - LiveDataSpeedBoost
		}

		cmd := tick()
		cmds = append(cmds, cmd)

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

		label := lipgloss.NewStyle().
			Bold(true).Render(fmt.Sprintf(`#%d activity`, m.list.Index()+1))

		var extraLabel string
		if m.showLiveData {
			extraLabel = lipgloss.NewStyle().
				PaddingLeft(1).PaddingRight(1).
				Bold(true).
				Reverse(true).
				Render("live")
		}

		detailsView += lipgloss.NewStyle().
			Bold(true).
			PaddingRight(4).
			Border(lipgloss.ASCIIBorder(), false, false, true, false).
			Render(label + " " + extraLabel)

		var playLabel string
		if m.showLiveData {
			playLabel = "paused"
			if m.playLiveData {
				playLabel = fmt.Sprintf("playing (speed %dx)", int(m.liveDataSpeed))
			}
		}

		detailsView += br
		detailsView += lipgloss.NewStyle().Italic(true).MarginBottom(1).Render(playLabel)

		const BAR_WIDTH = 50
		var col1 = lipgloss.NewStyle().Width(BAR_WIDTH / 2).Render
		var col2 = lipgloss.NewStyle().Width(BAR_WIDTH / 2).Align(lipgloss.Right).Render
		var th = lipgloss.NewStyle().Bold(true).Render

		if act, ok := item.(*common.Activity); ok {

			totalDistance := act.TotalDistance().Format()

			var rows [][]string
			if ad, ok := asyncdata.Success(act.Data); ok {

				currentRecord := ad.Records[act.RecordIndex()]

				rps := act.RPS()
				noRecordsText := fmt.Sprintf(`%d (%.1frps)`, ad.NoRecords(), rps)
				noSessionsText := fmt.Sprintf(`%d`, ad.NoSessions)

				b0 := BarEmpty
				b1 := BarEmptyHalf
				b2 := BarFullHalf

				if m.showLiveData {
					timeTxt := currentRecord.Time.FormatDate() + " " + currentRecord.Time.FormatHhMmSs()

					distanceTxt := col1("start") +
						col2("finish")
					distanceBar := HorizontalBar(
						float64(currentRecord.Distance.Value),
						b1,
						float64(ad.TotalDistance.Value),
						b0,
						BAR_WIDTH)

					currentDuration := TimeToDuration(ad.StartTime(), currentRecord.Time)
					finalDuration := TimeToDuration(ad.StartTime(), ad.FinishTime())
					durationTxt := col1("start") +
						col2("finish")
					durationBar := HorizontalBar(
						float64(currentDuration.Value),
						b1,
						float64(finalDuration.Value),
						b0,
						BAR_WIDTH)

					speedTxt := col1("min") + col2("max")
					speedBarTxt := common.NoDataText
					if currentRecord.Speed != nil {
						speedBarTxt = currentRecord.Speed.Format()
					}
					speedBar := HorizontalBar(0, b1, 0, b0, BAR_WIDTH)
					if ad.Speed.Max != nil && currentRecord.Speed != nil {
						speedBar = HorizontalBar(
							float64(currentRecord.Speed.Value),
							b1,
							float64(ad.Speed.Max.Value),
							b0,
							BAR_WIDTH)
					}

					altitudeTxt := col1("min") + col2("max")
					altitudeBarTxt := common.NoDataText
					if currentRecord.Altitude != nil {
						altitudeBarTxt = currentRecord.Altitude.Format()
					}
					altitudeBar := HorizontalBar(0, b1, 0, b0, BAR_WIDTH)
					if ad.Altitude.Min != nil && ad.Altitude.Max != nil && currentRecord.Altitude != nil {
						altitudeBar = HorizontalBarWithRange(
							float64(currentRecord.Altitude.Value),
							b1,
							ad.Altitude.Min.Value,
							float64(ad.Altitude.Max.Value),
							b0,
							BAR_WIDTH)
					}

					temperatureTxt := col1("min") + col2("max")
					temperatureBarTxt := common.NoDataText
					if currentRecord.Temperature != nil {
						temperatureBarTxt = currentRecord.Temperature.Format()
					}
					temperatureBar := HorizontalBar(0, b1, 0, b0, BAR_WIDTH)
					if ad.Temperature.Min != nil && ad.Temperature.Max != nil && currentRecord.Temperature != nil {
						temperatureBar = HorizontalBarWithRange(
							float64(currentRecord.Temperature.Value),
							b1,
							float64(ad.Temperature.Min.Value),
							float64(ad.Temperature.Max.Value),
							b0,
							BAR_WIDTH)
					}

					gpsTxt := col1("best") + col2("worst")
					gpsBarTxt := common.NoDataText
					if currentRecord.GpsAccuracy != nil {
						gpsBarTxt = currentRecord.GpsAccuracy.Format()
					}
					gpsBar := HorizontalBar(0, b1, 0, b0, BAR_WIDTH)
					if ad.GpsAccuracy.Min != nil && ad.GpsAccuracy.Max != nil && currentRecord.GpsAccuracy != nil {
						gpsBar = HorizontalBarWithRange(
							float64(currentRecord.GpsAccuracy.Value),
							b1,
							float64(ad.GpsAccuracy.Min.Value),
							float64(ad.GpsAccuracy.Max.Value),
							b0,
							BAR_WIDTH)
					}

					heartrateTxt := col1("min") + col2("max")
					heartrateBarTxt := common.NoDataText
					if currentRecord.Heartrate != nil {
						heartrateBarTxt = currentRecord.Heartrate.Format()
					}
					heartrateBar := HorizontalBar(0, b1, 0, b0, BAR_WIDTH)
					if ad.Heartrate.Min != nil && ad.Heartrate.Max != nil && currentRecord.Heartrate != nil {
						heartrateBar = HorizontalBarWithRange(
							float64(currentRecord.Heartrate.Value),
							b1,
							float64(ad.Heartrate.Min.Value),
							float64(ad.Heartrate.Max.Value),
							b0,
							BAR_WIDTH)
					}

					rows = [][]string{
						{th("time"), timeTxt},
						{th("distance"), distanceTxt},
						{currentRecord.Distance.Format3(), distanceBar},
						{th("duration"), durationTxt},
						{currentDuration.Format(), durationBar},
						{th("speed"), speedTxt},
						{speedBarTxt, speedBar},
						{th("altitude"), altitudeTxt},
						{altitudeBarTxt, altitudeBar},
						{th("temperature"), temperatureTxt},
						{temperatureBarTxt, temperatureBar},
						{th("gps accuracy"), gpsTxt},
						{gpsBarTxt, gpsBar},
						{th("♥ rate"), heartrateTxt},
						{heartrateBarTxt, heartrateBar},
						{th("sessions"), noSessionsText},
						{th("record"), fmt.Sprint(act.RecordIndex()+1) + " of " + noRecordsText},
					}
				} else {
					dateTxt := ad.StartTime().Format() + "-" + ad.FinishTime().FormatHhMm()
					durationTxt := common.NoDataText
					durationBar := HorizontalBar(0, b1, 0, b0, BAR_WIDTH)
					if ad.Duration.Active != nil && ad.Duration.Pause != nil {
						durationTxt = col1(ad.Duration.Active.Format())
						pauseTxt := "pause " + ad.Duration.Pause.Format()
						if ad.Duration.Pause.Value <= 0 {
							pauseTxt = "(no pause)"
						}

						durationTxt += col2(pauseTxt)
						durationBar = HorizontalStackedBar(
							float64(ad.Duration.Active.Value),
							b1,
							float64(ad.Duration.Pause.Value),
							b2,
							BAR_WIDTH)
					}

					speedTxt := common.NoDataText
					speedBar := HorizontalBar(0, b1, 0, b0, BAR_WIDTH)
					if ad.Speed.Avg != nil && ad.Speed.Max != nil {
						speedTxt = col1("⌀ "+ad.Speed.Avg.Format()) + col2("max "+ad.Speed.Max.Format())
						speedBar = HorizontalBar(
							float64(ad.Speed.Avg.Value),
							b1,
							float64(ad.Speed.Max.Value),
							b0,
							BAR_WIDTH)

					}

					elevationTxt := common.NoDataText
					elevationBar := HorizontalBar(0, b1, 0, b0, BAR_WIDTH)
					if ad.Elevation.Ascents != nil && ad.Elevation.Descents != nil {
						elevationTxt = col1(arrowTop+" "+ad.Elevation.Ascents.Format()) +
							col2(arrowDown+" "+ad.Elevation.Descents.Format())
						elevationBar = HorizontalStackedBar(
							float64(ad.Elevation.Ascents.Value),
							b1,
							float64(ad.Elevation.Descents.Value),
							b2,
							BAR_WIDTH)
					}

					temperatureTxt := common.NoDataText
					temperatureBar := HorizontalBar(0, b1, 0, b0, BAR_WIDTH)
					if ad.Temperature.Avg != nil && ad.Temperature.Max != nil {
						temperatureTxt = col1("⌀ "+ad.Temperature.Avg.Format()) +
							col2("max "+ad.Temperature.Max.Format())
						temperatureBar = HorizontalBar(
							float64(ad.Temperature.Avg.Value),
							b1,
							float64(ad.Temperature.Max.Value),
							b0,
							BAR_WIDTH)
					}

					gpsTxt := common.NoDataText
					gpsBar := HorizontalBar(0, b1, 0, b0, BAR_WIDTH)
					if ad.GpsAccuracy.Avg != nil && ad.GpsAccuracy.Max != nil {
						gpsTxt = col1("⌀ "+ad.GpsAccuracy.Avg.Format()) +
							col2("worse "+ad.GpsAccuracy.Max.Format())
						gpsBar = HorizontalBar(
							float64(ad.GpsAccuracy.Avg.Value),
							b1,
							float64(ad.GpsAccuracy.Max.Value),
							b0,
							BAR_WIDTH)
					}

					heartrateTxt := common.NoDataText
					heartrateBar := HorizontalBar(0, b1, 0, b0, BAR_WIDTH)
					if ad.Heartrate.Avg != nil && ad.Heartrate.Max != nil {
						heartrateTxt = col1("⌀ "+ad.Heartrate.Avg.Format()) +
							col2("max "+ad.Heartrate.Max.Format())
						heartrateBar = HorizontalBar(
							float64(ad.Heartrate.Avg.Value),
							b1,
							float64(ad.Heartrate.Max.Value),
							b0,
							BAR_WIDTH)
					}

					rows = [][]string{
						{th("date"), dateTxt},
						{th("distance"), totalDistance},
						{th("active"), durationTxt},
						{"", durationBar},
						{th("speed"), speedTxt},
						{"", speedBar},
						{th("elevation"), elevationTxt},
						{"", elevationBar},
						{th("temperature"), temperatureTxt},
						{"", temperatureBar},
						{th("gps accuracy"), gpsTxt},
						{"", gpsBar},
						{th("♥ rate"), heartrateTxt},
						{"", heartrateBar},
						{th("sessions"), noSessionsText},
						{th("records"), noRecordsText},
					}
				}

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
						return lipgloss.NewStyle().PaddingRight(2)
					case !m.showLiveData && (row == 1 || row == 13):
						return lipgloss.NewStyle().MarginBottom(2)
					case m.showLiveData && (row == 2 || row == 14):
						return lipgloss.NewStyle().MarginBottom(2)
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

	// title -> filtered
	if m.list.IsFiltered() && noVisibleActs != noActs {
		labelNoActs := fmt.Sprintf("%d of %d", noVisibleActs, len(m.list.Items()))
		m.list.Title = lipgloss.NewStyle().Bold(false).Italic(true).Render(labelNoActs)
	} else
	// title -> all
	{
		title := lipgloss.NewStyle().Bold(true).Render("All ")
		if ActivitiesParsing(m.activities) {
			title += lipgloss.NewStyle().Bold(false).Italic(true).Render("importing")
		}
		m.list.Title = title
	}

	view := m.list.View()

	sortLabel := "sorted by "
	switch m.actsSort {
	case DistanceAsc:
		sortLabel += "dist. " + arrowTop
	case DistanceDesc:
		sortLabel += "dist. " + arrowDown
	case TimeAsc:
		sortLabel += "time " + arrowTop
	case TimeDesc:
		sortLabel += "time " + arrowDown
	}

	// empty label for a single item
	if len(m.list.VisibleItems()) <= 1 {
		sortLabel = " "
	}

	view += lipgloss.NewStyle().PaddingLeft(2).
		MarginTop(2).
		MarginBottom(1).
		Italic(true).
		Render(sortLabel)

	return view
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
		filterCol2 := "[/]start"
		if m.list.SettingFilter() {
			if len(m.list.FilterValue()) > 0 {
				filterCol2 = "[ENTER]apply"
			} else {
				filterCol2 = "[ESC]cancel"
			}
		}
		filterCol3 := ""
		if m.list.IsFiltered() {
			filterCol3 = "[ESC]clear"
		}
		if m.list.SettingFilter() {
			if len(m.list.FilterValue()) > 0 {
				filterCol3 = "[ESC]skip"
			}
		}

		col := lipgloss.NewStyle().PaddingRight(3).Render

		liveDataTxt := col("[l]show")
		if m.showLiveData {
			liveDataTxt = col("[l]hide")
			if m.playLiveData {
				liveDataTxt += col("[space]stop")
				liveDataTxt += col("[1-9]1-9x")
				liveDataTxt += col("[0]10x")
				liveDataTxt += col("[→]+1x")
				liveDataTxt += col("[←]-1x")
				liveDataTxt += col("[^+→]boost")
			} else {
				liveDataTxt += col("[space]play")
				liveDataTxt += col("[→]next")
				liveDataTxt += col("[^→]ffw")
				liveDataTxt += col("[←]prev.")
				liveDataTxt += col("[^←]rwd")
			}
			liveDataTxt += col("[r]eset")
			liveDataTxt += col("[^r]eset all")
		}

		sortTxt := col("[^t]ime") + col("[^d]uration")

		listTxt := col("["+arrowTop+"]up") +
			col("["+arrowDown+"]down") +
			col("[g]first", "[G]last")

		filterTxt := col(filterCol2) + col(filterCol3)

		rows := [][]string{
			{"list", listTxt},
			{"sort", sortTxt},
			{"filter", filterTxt},
			{"live data", liveDataTxt},
		}
		table := table.New().
			Rows(rows...).
			Border(lipgloss.Border{}).
			StyleFunc(func(row, col int) lipgloss.Style {
				switch col {
				case 0:
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
	m.list.SetHeight(contentHeight - 3)     // offset for custom footer below list
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
			resultCh <- parseFileResultMsg{act}
			close(resultCh)
		}()
		// return result msg
		return <-resultCh
	}
}
