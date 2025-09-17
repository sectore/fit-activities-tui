package common

import (
	"fmt"
	"sort"
	"strings"

	"time"

	"github.com/sectore/fit-activities-tui/internal/asyncdata"
)

type Time struct{ Value time.Time }

func NewTime(value time.Time) Time {
	return Time{Value: value}
}

func (t Time) Format() string {
	return t.Value.Format("02.01.06 15:04")
}

func (t Time) FormatDate() string {
	return t.Value.Format("02.01.06")
}

func (t Time) FormatHhMm() string {
	return t.Value.Format("15:04")
}

func (t Time) FormatHhMmSs() string {
	return t.Value.Format("15:04:05")
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

// Duration in milliseconds
type Duration struct{ Value uint32 }

// Creates a new `Duration` expecting a value in milliseconds
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
		return fmt.Sprintf("%ds", seconds)
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

type Altitude struct{ Value float64 }

func NewAltitude(value float64) Altitude {
	return Altitude{Value: value}
}

func (a Altitude) Format() string {
	if a.Value == 0 {
		return "0m"
	}
	return fmt.Sprintf("%.fm", a.Value)
}

type AltitudeStats struct {
	Min Altitude
	Max Altitude
}

type Distance struct{ Value uint32 }

func NewDistance(value uint32) Distance {
	return Distance{Value: value}
}

func (d Distance) Format() string {
	return d.format(1)
}

func (d Distance) Format2() string {
	return d.format(2)
}

func (d Distance) Format3() string {
	return d.format(3)
}

// Internal function to format `Duration` by given decimal numbers
func (d Distance) format(decimal int) string {
	var meters = d.Value / 100
	if meters >= 1000 {
		km := float64(meters) / 1000
		var d string
		switch decimal {
		case 1:
			d = fmt.Sprintf("%.1f", km)
			d = strings.TrimRight(d, "0")
			d = strings.TrimRight(d, ".")
		case 2:
			d = fmt.Sprintf("%.2f", km)
		case 3:
			d = fmt.Sprintf("%.3f", km)
		default:
			d = fmt.Sprintf("%d", int(km))
		}
		return d + "km"
	} else {
		return fmt.Sprintf("%dm", meters)
	}
}

type RecordData struct {
	Time        Time
	Distance    Distance
	Speed       Speed
	Temperature Temperature
	GpsAccuracy GpsAccuracy
	Altitude    Altitude
}

type ActivityData struct {
	Duration      DurationStats
	TotalDistance Distance
	Speed         SpeedStats
	Temperature   TemperatureStats
	Elevation     ElevationStats
	Altitude      AltitudeStats
	NoSessions    uint32
	Records       []RecordData
	GpsAccuracy   GpsAccuracyStats
}

func (ad ActivityData) NoRecords() int {
	return len(ad.Records)
}

func (ad ActivityData) StartTime() Time {
	return ad.Records[0].Time
}

func (ad ActivityData) FinishTime() Time {
	last := max(ad.NoRecords()-1, 0)
	return ad.Records[last].Time
}

type ActivityAD = asyncdata.AsyncData[error, ActivityData]

type Activity struct {
	Path string
	/// index of current selected `Record`
	recordIndex int
	Data        ActivityAD
}

func (act Activity) FilterValue() string {
	var value string
	if data, ok := asyncdata.Success(act.Data); ok {
		value = (*data).StartTime().Format()
	}
	return value

}

func (act Activity) Title() string {
	var title string
	if data, ok := asyncdata.Success(act.Data); ok {
		title = (*data).StartTime().Format()
	}
	return title
}

func (act Activity) Description() string {
	return act.TotalDistance().Format()
}

func (act Activity) TotalDistance() Distance {
	if data, ok := asyncdata.Success(act.Data); ok {
		return (*data).TotalDistance
	}
	return NewDistance(0)
}

// default time: January 1, 1970 UTC
var defaultTime = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)

func (act Activity) StartTime() Time {
	if data, ok := asyncdata.Success(act.Data); ok {
		return (*data).StartTime()
	}
	return NewTime(defaultTime)
}

func (act Activity) FinishTime() Time {
	if data, ok := asyncdata.Success(act.Data); ok {
		return (*data).FinishTime()
	}
	return NewTime(defaultTime)
}

func (act Activity) GetTotalDuration() Duration {
	total := NewDuration(0)
	if data, ok := asyncdata.Success(act.Data); ok {
		total.Value += (*data).Duration.Total.Value
	}
	return total
}

func (act Activity) RecordIndex() int {
	return act.recordIndex
}

func (act *Activity) CountRecordIndex() bool {
	if data, ok := asyncdata.Success(act.Data); ok {
		if act.recordIndex < len(data.Records)-1 {
			act.recordIndex += 1
			return true
		}
	}
	return false
}

func (act *Activity) ResetRecordIndex() {
	act.recordIndex = 0
}

type Activities = []*Activity // pointer slice to mutate values of `Activity`

type SortBy func(act1, act2 *Activity) bool

func (by SortBy) Sort(acts Activities) {
	as := &actSorter{
		acts: acts,
		by:   by,
	}
	sort.Sort(as)
}

func (by SortBy) Reverse(acts Activities) {
	as := &actSorter{
		acts: acts,
		by:   by,
	}
	sort.Sort(sort.Reverse(as))
}

type actSorter struct {
	acts Activities
	by   func(act1, act2 *Activity) bool
}

func (s *actSorter) Len() int {
	return len(s.acts)
}

func (s *actSorter) Swap(i, j int) {
	s.acts[i], s.acts[j] = s.acts[j], s.acts[i]
}

func (s *actSorter) Less(i, j int) bool {
	return s.by(s.acts[i], s.acts[j])
}

var SortByDistance = func(act1, act2 *Activity) bool {
	return act1.TotalDistance().Value < act2.TotalDistance().Value
}

var SortByTime = func(act1, act2 *Activity) bool {
	return act1.StartTime().Value.Before(act2.StartTime().Value)
}
