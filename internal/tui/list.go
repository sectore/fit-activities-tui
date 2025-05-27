package tui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sectore/fit-sum-tui/internal/asyncdata"
	"github.com/sectore/fit-sum-tui/internal/common"
)

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
