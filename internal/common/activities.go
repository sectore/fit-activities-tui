package common

import (
	"fmt"
	"time"

	"github.com/sectore/fit-activities-tui/internal/asyncdata"
)

type Temperature = int8
type Temperatures = []Temperature

type Ascent = uint16
type Ascents = []Ascent

type Descent = uint16
type Descents = []Descent

type GpsAccuracy struct{ Value float32 }

func NewGpsAccuracy(value float32) GpsAccuracy {
	return GpsAccuracy{Value: value}
}

func (ga GpsAccuracy) Format() string {
	return fmt.Sprintf("%.1fm", ga.Value)
}

type GpsAccuracyStat struct {
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

type ActivityData struct {
	LocalTime      time.Time
	TotalTime      uint32
	TotalDistances []uint32
	Speed          SpeedStats
	Temperatures   Temperatures
	Descents       Descents
	Ascents        Ascents
	NoSessions     uint32
	NoRecords      uint32
	GpsAccuracy    GpsAccuracyStat
}

func (act ActivityData) TotalDistance() uint32 {
	value := uint32(0)
	for _, d := range act.TotalDistances {
		value += d
	}
	return value
}

func (act ActivityData) TotalDescant() Descent {
	l := len(act.Descents)
	value := Descent(0)
	for _, d := range act.Descents {
		value += d
	}
	return value / uint16(l)
}

func (act ActivityData) TotalAscant() Ascent {
	value := Ascent(0)
	l := len(act.Ascents)
	for _, d := range act.Ascents {
		value += d
	}
	return value / uint16(l)
}

type TemperatureStats struct {
	Avg, Min, Max Temperature
}

func (act ActivityData) Temperature() TemperatureStats {
	l := len(act.Temperatures)

	if l == 0 {
		return TemperatureStats{0, 0, 0}
	}

	total := 0
	count := 0
	min := act.Temperatures[0]
	max := int8(0)

	for _, t := range act.Temperatures {
		// for any reason Wahoo ELMNT counts 127 at start
		if t < 100 {
			count += 1
			total += int(t)
			if t < min {
				min = t
			}
			if max < t {
				max = t
			}
		}
	}

	avg := total / count
	return TemperatureStats{Avg: int8(avg), Max: max, Min: min}
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
