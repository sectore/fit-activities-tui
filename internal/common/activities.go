package common

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sectore/fit-activities-tui/internal/asyncdata"
)

type Time struct{ Value time.Time }

func NewTime(value time.Time) Time {
	return Time{Value: value}
}

func (t Time) Format() string {
	// return data.LocalTime.Format("2006-01-02 15:04")
	return t.Value.Format("02.01.06 15:04")
}

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

type Duration struct{ Value uint32 }

func NewDuration(value uint32) Duration {
	return Duration{Value: value}
}

type DurationStats struct {
	Total  Duration
	Active Duration
	Pause  Duration
}

func (time Duration) Format() string {
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
	StartTime     Time
	Duration      DurationStats
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
		value = data.StartTime.Format()
	}
	return value

}

func (act Activity) Title() string {
	var title string
	if data, ok := asyncdata.Success(act.Data); ok {
		title = data.StartTime.Format()
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

func (act Activity) GetTotalDuration() Duration {
	total := NewDuration(0)
	if data, ok := asyncdata.Success(act.Data); ok {
		total.Value += data.Duration.Total.Value
	}
	return total
}

type Activities = []Activity
