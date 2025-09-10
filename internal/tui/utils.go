package tui

import (
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
		if act, ok := item.(common.Activity); ok {
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

func HorizontalStackedBar(value1 float32, value1Block string, value2 float32, value2Block string, maxBlocks int) string {
	total := value1 + value2
	value2Percent := value2 * float32(maxBlocks) / total
	// integer is needed to count blocks
	noValue2Blocks := int(value2Percent)
	// adjust to still show small `pauseValue`s < 1
	if noValue2Blocks == 0 && value2 > 0 {
		noValue2Blocks = 1
	}
	// other blocks are blocks for `value1`
	noValue1Blocks := maxBlocks - noValue2Blocks
	return strings.Repeat(value1Block, noValue1Blocks) +
		strings.Repeat(value2Block, noValue2Blocks)
}
