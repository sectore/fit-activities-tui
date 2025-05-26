package common

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sectore/fit-sum-tui/internal/asyncdata"
)

type ActivityData struct {
	LocalTime      time.Time
	TotalTime      uint32
	TotalDistances []uint32
}

type ActivityAD struct {
	asyncdata.AsyncData[error, ActivityData]
}

func (ad ActivityAD) FilterValue() string {
	if act, ok := asyncdata.Success[error, ActivityData](ad.AsyncData); ok {
		return act.FormatLocalTime()
	}
	return ""
}

func (ad ActivityAD) Title() string {
	if act, ok := asyncdata.Success[error, ActivityData](ad.AsyncData); ok {
		return act.FormatLocalTime()
	}
	// if ok := asyncdata.NotAsked[error, ActivityData](ad.AsyncData); ok {
	// 	return "loading"
	// }
	return ""
}

func (ad ActivityAD) Description() string {
	if act, ok := asyncdata.Success[error, ActivityData](ad.AsyncData); ok {
		return act.FormatTotalDistance()
	}
	return ""
}

// func (act ActivityAD) Render() string      { return "" }

type Activity struct {
	Path string
	Data ActivityAD
}

func (data ActivityData) GetTotalDistance() uint32 {
	var dist uint32 = 0
	for _, d := range data.TotalDistances {
		dist += d
	}
	return dist
}

func (data ActivityData) FormatTotalDistance() string {
	var meters = data.GetTotalDistance() / 100
	if meters >= 1000 {
		km := float64(meters) / 1000
		formatted := fmt.Sprintf("%.1f", km)
		formatted = strings.TrimRight(formatted, "0")
		formatted = strings.TrimRight(formatted, ".")
		return formatted + "km"
	} else {
		return fmt.Sprintf("%dm", meters)
	}
}

func (data ActivityData) FormatLocalTime() string {
	// return data.LocalTime.Format("2006-01-02 15:04")
	return data.LocalTime.Format("02.01.06 15:04")

}

func (data ActivityData) FormatTotalTime() string {
	var s string
	seconds := int(data.TotalTime / 1000)
	if seconds < 60 {
		// Format as ss
		s = strconv.Itoa(seconds)
	} else if seconds < 3600 {
		// Format as mm:ss
		minutes := seconds / 60
		remainingSeconds := seconds % 60
		s = fmt.Sprintf("%02dm %02ds", minutes, remainingSeconds)
	} else if seconds < 86400 {
		// Format as hh:mm:ss
		hours := seconds / 3600
		remainingSeconds := seconds % 3600
		minutes := remainingSeconds / 60
		seconds = remainingSeconds % 60
		s = fmt.Sprintf("%02dh %02dm %02ds", hours, minutes, seconds)
	} else {
		// Format as dd:hh:mm:ss
		days := seconds / 86400
		remainingSeconds := seconds % 86400
		hours := remainingSeconds / 3600
		remainingSeconds = remainingSeconds % 3600
		minutes := remainingSeconds / 60
		seconds = remainingSeconds % 60
		s = fmt.Sprintf("%dd %02dh %02dm %02ds", days, hours, minutes, seconds)
	}

	return s
}

type Activities = []Activity
