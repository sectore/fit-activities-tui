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
	return asyncdata.NewLoading[error, ActivityData](prevData)
}

func ActivityFailure(err error) ActivityAD {
	return asyncdata.NewFailure[error, ActivityData](err)
}

func ActivitySuccess(data ActivityData) ActivityAD {
	return asyncdata.NewSuccess[error, ActivityData](data)
}

type Activity struct {
	Path     string
	selected bool
	Data     ActivityAD
}

func (act *Activity) Toggle() {
	act.selected = !act.selected
}

func (act Activity) IsSelected() bool {
	return act.selected
}

type Activities struct {
	index uint
	inner []Activity
}

func NewActivities(acts []Activity) Activities {
	return Activities{inner: acts, index: 0}
}

func (acts Activities) All() []Activity {
	return acts.inner
}

func (acts *Activities) Next() (*Activity, bool) {
	var l = uint(len(acts.inner))
	// empty lists
	if l == 0 {
		return nil, false
	}
	acts.index = (acts.index + 1) % l
	return &acts.inner[acts.index], true
}

func (acts *Activities) Prev() (*Activity, bool) {
	var l = uint(len(acts.inner))
	// empty lists
	if l == 0 {
		return nil, false
	}
	acts.index = (acts.index + l - 1) % l
	return &acts.inner[acts.index], true
}

func (acts Activities) IsLastIndex() bool {
	var l = uint(len(acts.inner))
	if l == 0 {
		return true
	}
	return acts.index >= l-1
}

func (acts *Activities) FirstIndex() {
	acts.index = 0
}

func (acts Activities) IsFirstIndex() bool {
	return acts.index == 0
}

func (acts Activities) CurrentIndex() uint {
	return acts.index
}

func (acts *Activities) CurrentAct() (*Activity, bool) {
	// empty lists
	if len(acts.inner) == 0 {
		return nil, false
	}
	return &acts.inner[acts.index], true
}
