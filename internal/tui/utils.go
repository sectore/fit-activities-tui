package tui

import (
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

func ActivitiesTotalDistances(acts common.Activities) uint32 {
	var dist uint32
	for _, act := range acts {
		dist += act.TotalDistance()
	}
	return dist
}

func ActivitiesTotalTime(acts common.Activities) uint32 {
	var total uint32
	for _, act := range acts {
		total += act.GetTotalTime()
	}
	return total
}

func ListItemsToActivities(items []list.Item) common.Activities {
	// acts := make(common.Activities, len(items))
	var acts common.Activities
	for _, item := range items {
		if act, ok := item.(*common.Activity); ok {
			acts = append(acts, *act)
		}
	}
	return acts
}
