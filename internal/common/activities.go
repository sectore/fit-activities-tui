package common

import (
	"time"

	"github.com/sectore/fit-activities-tui/internal/asyncdata"
)

type ActivityData struct {
	LocalTime      time.Time
	TotalTime      uint32
	TotalDistances []uint32
}

func (act ActivityData) TotalDistance() uint32 {
	var value uint32 = 0
	for _, d := range act.TotalDistances {
		value += d
	}
	return value
}

type ActivityAD = asyncdata.AsyncData[error, ActivityData]

type Activity struct {
	Path string
	Data ActivityAD
}

func (act Activity) FilterValue() string {
	var value string
	if data, ok := asyncdata.Success[error, ActivityData](act.Data); ok {
		value = FormatLocalTime(data.LocalTime)
	}
	return value

}

func (act Activity) Title() string {
	var title string
	if data, ok := asyncdata.Success[error, ActivityData](act.Data); ok {
		title = FormatLocalTime(data.LocalTime)
	}
	return title
}

func (act Activity) Description() string {
	return FormatTotalDistance(act.TotalDistance())
}

func (act Activity) TotalDistance() uint32 {
	var value uint32 = 0
	if data, ok := asyncdata.Success[error, ActivityData](act.Data); ok {
		value += data.TotalDistance()
	}
	return value
}

func (act Activity) GetTotalTime() uint32 {
	var value uint32 = 0
	if data, ok := asyncdata.Success[error, ActivityData](act.Data); ok {
		value += data.TotalTime
	}
	return value
}

type Activities = []Activity
