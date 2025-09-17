package tui

import (
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/sectore/fit-activities-tui/internal/asyncdata"
	"github.com/sectore/fit-activities-tui/internal/common"
)

func ActivitiesParsing(acts common.Activities) bool {
	for _, act := range acts {
		_, _, loading := asyncdata.Loading(act.Data)
		notAsked := asyncdata.NotAsked(act.Data)
		if loading || notAsked {
			return true
		}
	}
	return false
}

func ActivitiesParsed(acts common.Activities) int {
	count := 0
	for _, act := range acts {
		if _, ok := asyncdata.Success(act.Data); ok {
			count += 1
		}
	}
	return count
}

func ActivitiesFailures(acts common.Activities) int {
	count := 0
	for _, act := range acts {
		if _, ok := asyncdata.Failure(act.Data); ok {
			count += 1
		}
	}
	return count
}

func ActivitiesTotalDistances(acts common.Activities) common.Distance {
	dist := common.NewDistance(0)
	for _, act := range acts {
		dist.Value += act.TotalDistance().Value
	}
	return dist
}

func ActivitiesTotalDuration(acts common.Activities) common.Duration {
	total := common.NewDuration(0)
	for _, act := range acts {
		total.Value += act.GetTotalDuration().Value
	}
	return total
}

func ListItemsToActivities(items []list.Item) common.Activities {
	var acts common.Activities
	for _, item := range items {
		if act, ok := item.(*common.Activity); ok {
			acts = append(acts, act)
		}
	}
	return acts
}

func ActivitiesToListItems(acts common.Activities) []list.Item {
	items := make([]list.Item, len(acts))
	for i, act := range acts {
		items[i] = act
	}
	return items
}

func SortItems(items []list.Item, sort ActsSort) []list.Item {
	acts := ListItemsToActivities(items)
	switch sort {
	case DistanceAsc:
		common.SortBy(common.SortByDistance).Sort(acts)
	case DistanceDesc:
		common.SortBy(common.SortByDistance).Reverse(acts)
	case TimeAsc:
		common.SortBy(common.SortByTime).Sort(acts)
	case TimeDesc:
		common.SortBy(common.SortByTime).Reverse(acts)
	}
	return ActivitiesToListItems(acts)
}

func HorizontalStackedBar(value1 float64, value1Block string, value2 float64, value2Block string, maxBlocks int) string {
	total := value1 + value2
	value2Percent := value2 * float64(maxBlocks) / total
	// use rounding instead of truncation for better proportional representation
	noValue2Blocks := int(math.Round(value2Percent))
	// adjust `noValue2Blocks` to still show small values of `pauseValue` < 1
	if noValue2Blocks == 0 && value2 > 0 {
		noValue2Blocks = 1
	}
	// adjust `noValue1Blocks` to never be < 0
	// to avoid negative `Repeat` count
	noValue1Blocks := math.Max(float64(maxBlocks-noValue2Blocks), 0)
	return strings.Repeat(value1Block, int(noValue1Blocks)) +
		strings.Repeat(value2Block, noValue2Blocks)
}

func HorizontalBar(value float64, fgBlock string, maxValue float64, bgBlock string, maxBlocks int) string {
	maxBlocks_f64 := float64(maxBlocks)
	noValueBlocks := value * maxBlocks_f64 / maxValue
	// ensure noValueBlocks doesn't exceed maxBlocks
	if noValueBlocks > maxBlocks_f64 {
		noValueBlocks = maxBlocks_f64
	}

	// avoid negative `Repeat` count
	noValueBlocks = math.Max(float64(noValueBlocks), 0)
	// re-adjust to show a block asap
	if int(noValueBlocks) == 0 && value > 0 {
		noValueBlocks = 1
	}

	// convert to integer for foreground blocks
	fgBlocks := int(noValueBlocks)
	// calculate background blocks to ensure total equals maxBlocks
	bgBlocks := maxBlocks - fgBlocks

	return strings.Repeat(fgBlock, fgBlocks) +
		strings.Repeat(bgBlock, bgBlocks)
}

func TimeToDuration(start common.Time, end common.Time) common.Duration {
	seconds := end.Value.Unix() - start.Value.Unix()
	seconds = int64(math.Max(float64(seconds), 0))
	milliseconds := uint32(seconds * 1000)
	return common.NewDuration(milliseconds)
}
