package common

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sectore/fit-activities-tui/internal/asyncdata"
)

type Temperature struct{ Value float32 }

func NewTemperature(value float32) Temperature {
	return Temperature{Value: value}
}

func (t Temperature) Format() string {
	return fmt.Sprintf("%0.f°C", t.Value)
}

type TemperatureStats struct {
	Avg, Min, Max Temperature
}

type GpsAccuracy struct{ Value float32 }

func NewGpsAccuracy(value float32) GpsAccuracy {
	return GpsAccuracy{Value: value}
}

func (ga GpsAccuracy) Format() string {
	return fmt.Sprintf("%.0fm", ga.Value)
}

type GpsAccuracyStats struct {
	Avg, Min, Max GpsAccuracy
}

type Speed struct{ Value float32 }

func NewSpeed(value float32) Speed {
	return Speed{Value: value}
}

func (s Speed) Format() string {
	return fmt.Sprintf("%.1fkm/h", s.Value*3.6/1000)
}

type SpeedStats struct {
	Avg, Max Speed
}

type Time struct{ Value uint32 }

func NewTime(value uint32) Time {
	return Time{Value: value}
}

type TimeStats struct {
	Active Time
	Total  Time
	Pause  Time
}

func (time Time) Format() string {
	seconds := time.Value / 1000
	if seconds < 60 {
		return strconv.Itoa(int(seconds))
	} else if seconds < 3600 {
		minutes := seconds / 60
		remainingSeconds := seconds % 60
		return fmt.Sprintf("%dm %ds", minutes, remainingSeconds)
	} else if seconds < 86400 {
		hours := seconds / 3600
		remainingSeconds := seconds % 3600
		minutes := remainingSeconds / 60
		seconds = remainingSeconds % 60
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else {
		days := seconds / 86400
		remainingSeconds := seconds % 86400
		hours := remainingSeconds / 3600
		remainingSeconds = remainingSeconds % 3600
		minutes := remainingSeconds / 60
		seconds = remainingSeconds % 60
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	}
}

type Elevation struct{ Value uint16 }

func NewElevation(value uint16) Elevation {
	return Elevation{Value: value}
}

func (e Elevation) Format() string {
	return fmt.Sprintf("%dm", e.Value)
}

type ElevationStats struct {
	Descents Elevation
	Ascents  Elevation
}

type Distance struct{ Value uint32 }

func NewDistance(value uint32) Distance {
	return Distance{Value: value}
}

func (d Distance) Format() string {
	var meters = d.Value / 100
	if meters >= 1000 {
		km := float64(meters) / 1000
		d := fmt.Sprintf("%.1f", km)
		d = strings.TrimRight(d, "0")
		d = strings.TrimRight(d, ".")
		return d + "km"
	} else {
		return fmt.Sprintf("%dm", meters)
	}
}

type ActivityData struct {
	LocalTime     time.Time
	Time          TimeStats
	TotalDistance Distance
	Speed         SpeedStats
	Temperature   TemperatureStats
	Elevation     ElevationStats
	NoSessions    uint32
	NoRecords     uint32
	GpsAccuracy   GpsAccuracyStats
}

type ActivityAD = asyncdata.AsyncData[error, ActivityData]

type Activity struct {
	Path string
	Data ActivityAD
}

func (act Activity) FilterValue() string {
	var value string
	if data, ok := asyncdata.Success(act.Data); ok {
		value = FormatLocalTime(data.LocalTime)
	}
	return value

}

func (act Activity) Title() string {
	var title string
	if data, ok := asyncdata.Success(act.Data); ok {
		title = FormatLocalTime(data.LocalTime)
	}
	return title
}

func (act Activity) Description() string {
	return act.TotalDistance().Format()
}

func (act Activity) TotalDistance() Distance {
	if data, ok := asyncdata.Success(act.Data); ok {
		return data.TotalDistance
	}
	return NewDistance(0)
}

func (act Activity) GetTotalTime() Time {
	value := NewTime(0)
	if data, ok := asyncdata.Success(act.Data); ok {
		value.Value += data.Time.Total.Value
	}
	return value
}

type Activities = []Activity
