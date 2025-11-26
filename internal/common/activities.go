package common

import (
	"fmt"
	"sort"
	"strings"

	"time"

	"github.com/sectore/fit-activities-tui/internal/asyncdata"
)

const NoDataText = "no data"

// Helper to create a pointer to given value
func Ptr[T any](v T) *T {
	return &v
}

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

type Temperature struct{ Value int8 }

func NewTemperature(value int8) Temperature {
	return Temperature{Value: value}
}

func (t Temperature) Format() string {
	return fmt.Sprintf("%0d°C", t.Value)
}

type TemperatureStats struct {
	Avg, Min, Max *Temperature
}

type GpsAccuracy struct{ Value uint8 }

func NewGpsAccuracy(value uint8) GpsAccuracy {
	return GpsAccuracy{Value: value}
}

func (ga GpsAccuracy) Format() string {
	return fmt.Sprintf("%dm", ga.Value)
}

type GpsAccuracyStats struct {
	Avg, Min, Max *GpsAccuracy
}

type Speed struct{ Value float32 }

func NewSpeed(value float32) Speed {
	return Speed{Value: value}
}

func (s Speed) Format() string {
	return fmt.Sprintf("%.1fkm/h", s.Value*3.6/1000)
}

type SpeedStats struct {
	Avg, Max *Speed
}

// Duration in milliseconds
type Duration struct{ Value uint32 }

// Creates a new `Duration` expecting a value in milliseconds
func NewDuration(value uint32) Duration {
	return Duration{Value: value}
}

type DurationStats struct {
	Total, Active, Pause *Duration
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
	Descents, Ascents *Elevation
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
	Min, Max *Altitude
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

type Heartrate struct{ Value uint8 }

func NewHeartrate(value uint8) Heartrate {
	return Heartrate{Value: value}
}

func (hr Heartrate) Format() string {
	return fmt.Sprintf("%dbpm", hr.Value)
}

type HeartrateStats struct {
	Min, Max, Avg *Heartrate
}

type RecordData struct {
	Time        Time
	Distance    *Distance
	Speed       *Speed
	Temperature *Temperature
	GpsAccuracy *GpsAccuracy
	Altitude    *Altitude
	Heartrate   *Heartrate
}

type ActivityData struct {
	Duration      DurationStats
	TotalDistance *Distance
	Speed         SpeedStats
	Temperature   TemperatureStats
	Elevation     ElevationStats
	Altitude      AltitudeStats
	NoSessions    uint32
	Records       []RecordData
	GpsAccuracy   GpsAccuracyStats
	Heartrate     HeartrateStats
}

func (ad ActivityData) NoRecords() int {
	return len(ad.Records)
}

func (ad ActivityData) StartTime() *Time {
	if ad.NoRecords() == 0 {
		return nil
	}
	return &ad.Records[0].Time
}

func (ad ActivityData) FinishTime() *Time {
	if ad.NoRecords() == 0 {
		return nil
	}
	last := max(ad.NoRecords()-1, 0)
	return &ad.Records[last].Time
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
		if startTime := data.StartTime(); startTime != nil {
			value = startTime.Format()
		}
	}
	return value

}

func (act Activity) Title() string {
	var title string
	if data, ok := asyncdata.Success(act.Data); ok {
		if startTime := data.StartTime(); startTime != nil {
			title = startTime.Format()
		}
	}
	return title
}

func (act Activity) Description() string {
	return act.TotalDistance().Format()
}

func (act Activity) TotalDistance() Distance {
	if data, ok := asyncdata.Success(act.Data); ok {
		if data.TotalDistance != nil {
			return *data.TotalDistance
		}
	}
	return NewDistance(0)
}

// default time: January 1, 1970 UTC
var defaultTime = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)

func (act Activity) StartTime() *Time {
	if data, ok := asyncdata.Success(act.Data); ok {
		return data.StartTime()
	}
	return nil
}

func (act Activity) GetTotalDuration() Duration {
	total := NewDuration(0)
	if data, ok := asyncdata.Success(act.Data); ok {
		if data.Duration.Total != nil {
			total.Value += data.Duration.Total.Value
		}
	}
	return total
}

func (act Activity) RecordIndex() int {
	return act.recordIndex
}

func (act *Activity) CountRecordIndex(value int) bool {
	if data, ok := asyncdata.Success(act.Data); ok {
		length := len(data.Records)
		// empty records
		if length == 0 {
			act.recordIndex = 0
			return true
		}

		index := act.recordIndex + value
		max := length - 1

		// adjust to keep valid ranges
		if index > max {
			act.recordIndex = max
		} else if index < 0 {
			act.recordIndex = 0
		} else {
			act.recordIndex = index
		}

		return true
	}

	return false
}

func (act *Activity) ResetRecordIndex() {
	act.recordIndex = 0
}

// `Records` per second (RPS) based on total `Duration` and number of records
func (act *Activity) RPS() float64 {
	if data, ok := asyncdata.Success(act.Data); ok {
		if data.Duration.Total != nil {
			totalDurationMs := float64(data.Duration.Total.Value)
			recordCount := float64(len(data.Records))

			if totalDurationMs > 0 && recordCount > 0 {
				// ms to seconds
				totalDurationSeconds := totalDurationMs / 1000.0
				// records per second
				return recordCount / totalDurationSeconds
			}
		}
	}
	return 1.0 // fallback to 1 RPS
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
	startTime1Ptr := act1.StartTime()
	if startTime1Ptr == nil {
		t := NewTime(defaultTime)
		startTime1Ptr = Ptr(t)
	}
	startTime2Ptr := act2.StartTime()
	if startTime2Ptr == nil {
		t := NewTime(defaultTime)
		startTime2Ptr = Ptr(t)
	}
	return startTime1Ptr.Value.Before(startTime2Ptr.Value)
}
