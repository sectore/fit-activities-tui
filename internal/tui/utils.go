package tui

import (
	"github.com/sectore/fit-sum-tui/internal/asyncdata"
	"github.com/sectore/fit-sum-tui/internal/common"
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
