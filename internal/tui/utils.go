package tui

import (
	"github.com/sectore/fit-sum-tui/internal/asyncdata"
	"github.com/sectore/fit-sum-tui/internal/common"
)

func ActivitiesParsing(acts common.Activities) bool {
	for _, act := range acts.All() {
		_, _, loading := asyncdata.Loading(act.Data.AsyncData)
		notAsked := asyncdata.NotAsked(act.Data.AsyncData)
		if loading || notAsked {
			return true
		}
	}
	return false
}

func ActivitiesParsed(acts common.Activities) int {
	count := 0
	for _, act := range acts.All() {
		if _, ok := asyncdata.Success(act.Data.AsyncData); ok {
			count += 1
		}
	}
	return count
}

func ActivitiesFailures(acts common.Activities) int {
	count := 0
	for _, act := range acts.All() {
		if _, ok := asyncdata.Failure(act.Data.AsyncData); ok {
			count += 1
		}
	}
	return count
}
