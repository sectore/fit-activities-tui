package common

import (
	"time"

	"github.com/sectore/fit-sum-tui/internal/asyncdata"
)

type ActivityData struct {
	LocalTime time.Time
	TotalTime uint32
	Distances []uint32
}

type ActivityAD = asyncdata.AsyncData[error, ActivityData]

func ActivityLoading(prevData *ActivityData) ActivityAD {
	return asyncdata.Loading[error, ActivityData](prevData)
}

func ActivityFailure(err error) ActivityAD {
	return asyncdata.Failure[error, ActivityData](err)
}

func ActivitySuccess(data ActivityData) ActivityAD {
	return asyncdata.Success[error, ActivityData](data)
}

type Activity struct {
	Path     string
	Selected bool
	Data     ActivityAD
}

type Activities struct {
	inner []Activity
}

// Set / get `next` selected `Activity` from `Activities`
func (m Activities) Next() (*Activity, bool) {
	// ignore empty lists
	if len(m.inner) == 0 {
		return nil, false
	}

	for i, act := range m.inner {
		if act.Selected {
			// unselect current
			m.inner[i].Selected = false
			// select next
			next := (i + 1) % len(m.inner)
			m.inner[next].Selected = true
			return &m.inner[next], true
		}
	}
	// In case no `Activity` was selected, select first
	m.inner[0].Selected = true
	return &m.inner[0], true
}

// Set / get `prev` selected `Activity` from `Activities`
func (m Activities) Prev() (*Activity, bool) {
	// ignore empty lists
	if len(m.inner) == 0 {
		return nil, false
	}

	for i, act := range m.inner {
		if act.Selected {
			// unselect current
			m.inner[i].Selected = false
			// select prev
			prev := (i + len(m.inner) - 1) % len(m.inner)
			m.inner[prev].Selected = true
			return &m.inner[prev], true
		}
	}
	// In case no `Activity` was selected, select first
	m.inner[0].Selected = true
	return &m.inner[0], true
}

// Set / get `prev` selected `Activity` from `Activities`
func (m Activities) Current() (*Activity, bool) {
	// ignore empty lists
	if len(m.inner) == 0 {
		return nil, false
	}

	for _, act := range m.inner {
		if act.Selected {
			return &act, true
		}
	}

	return nil, true
}
