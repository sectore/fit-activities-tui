package fit

import (
	"errors"
	"os"

	"github.com/muktihari/fit/decoder"
	"github.com/muktihari/fit/profile/basetype"
	"github.com/muktihari/fit/profile/filedef"
	"github.com/muktihari/fit/profile/mesgdef"
	"github.com/sectore/fit-activities-tui/internal/common"
)

func parseSpeed(rs []*mesgdef.Record) common.SpeedStats {
	speed := common.SpeedStats{
		Avg: common.NewSpeed(0),
		Max: common.NewSpeed(0),
	}
	var count, total uint
	for _, r := range rs {
		value := r.Speed
		// don't count invalid values
		if value != basetype.Uint16Invalid {
			value_f := float32(r.Speed)
			// compare max
			if speed.Max.Value < value_f {
				speed.Max.Value = value_f
			}
			count += 1
			total += uint(value)
		}
	}
	// skip if no valid values found
	if count == 0 {
		return speed
	}

	avg := float32(total / count)
	speed.Avg.Value = avg
	return speed

}

func parseGpsAccurancies(rs []*mesgdef.Record) common.GpsAccuracyStat {
	// start w/ empty stats
	gpsAccurancy := common.GpsAccuracyStat{
		Avg: common.NewGpsAccuracy(0),
		Min: common.NewGpsAccuracy(0),
		Max: common.NewGpsAccuracy(0),
	}
	var sum, count uint
	for _, r := range rs {
		value := r.GpsAccuracy
		// don't count invalid values
		if value != basetype.Uint8Invalid {
			value_f := float32(value)
			// override default zero value
			if count == 0 {
				gpsAccurancy.Min.Value = value_f
			}
			// compare min
			if value_f < gpsAccurancy.Min.Value {
				gpsAccurancy.Min.Value = value_f
			}
			// compare max
			if value_f > gpsAccurancy.Max.Value {
				gpsAccurancy.Max.Value = value_f
			}
			count += 1
			sum += uint(value)
		}
	}
	// skip if no valid values found
	if count == 0 {
		return gpsAccurancy
	}

	gpsAccurancy.Avg.Value = float32(sum / count)
	return gpsAccurancy
}

func parseTime(ss []*mesgdef.Session) common.TimeStats {
	time := common.TimeStats{
		Total:  common.NewTime(0),
		Active: common.NewTime(0),
		Pause:  common.NewTime(0),
	}
	for _, s := range ss {
		total := s.TotalElapsedTime
		if total != basetype.Uint32Invalid {
			time.Total.Value += total
		}
		active := s.TotalTimerTime
		if active != basetype.Uint32Invalid {
			time.Active.Value += active
		}
	}
	time.Pause.Value = time.Total.Value - time.Active.Value
	return time
}
func parseElevation(ss []*mesgdef.Session) common.ElevationStats {
	elevation := common.ElevationStats{
		Ascents:  common.NewElevation(0),
		Descents: common.NewElevation(0),
	}

	for _, s := range ss {
		value := s.TotalAscent
		if value != basetype.Uint16Invalid {
			elevation.Ascents.Value += value
		}
		value = s.TotalDescent
		if value != basetype.Uint16Invalid {
			elevation.Descents.Value += value
		}
	}

	return elevation
}

func parseTemperature(rs []*mesgdef.Record) common.TemperatureStats {

	temperature := common.TemperatureStats{
		Min: common.NewTemperature(0),
		Max: common.NewTemperature(0),
		Avg: common.NewTemperature(0),
	}
	var sum, count uint
	for _, r := range rs {
		value := r.Temperature
		// don't count invalid values
		if value != basetype.Sint8Invalid {
			value_f := float32(value)
			// override default zero value
			if count == 0 {
				temperature.Min.Value = value_f
			}
			// compare min
			if value_f < temperature.Min.Value {
				temperature.Min.Value = value_f
			}
			// compare max
			if value_f > temperature.Max.Value {
				temperature.Max.Value = value_f
			}
			count += 1
			sum += uint(value)
		}
	}
	// skip if no valid values found
	if count == 0 {
		return temperature
	}

	temperature.Avg.Value = float32(sum / count)
	return temperature
}

func ParseFile(file string) (*common.ActivityData, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	lis := filedef.NewListener()
	defer lis.Close()

	dec := decoder.New(f,
		decoder.WithMesgListener(lis),
		decoder.WithBroadcastOnly(),
	)
	_, err = dec.Decode()
	if err != nil {
		return nil, err
	}

	act, ok := lis.File().(*filedef.Activity)
	if !ok {
		return nil, errors.New("expected an Activity")
	}

	noSessions := len(act.Sessions)
	noRecords := len(act.Records)

	totalDistance := common.NewDistance(0)
	for _, s := range act.Sessions {
		totalDistance.Value += s.TotalDistance
	}

	var activityData = common.ActivityData{
		LocalTime:     act.Activity.LocalTimestamp,
		Time:          parseTime(act.Sessions),
		TotalDistance: totalDistance,
		Temperature:   parseTemperature(act.Records),
		Speed:         parseSpeed(act.Records),
		Elevation:     parseElevation(act.Sessions),
		NoSessions:    uint32(noSessions),
		NoRecords:     uint32(noRecords),
		GpsAccuracy:   parseGpsAccurancies(act.Records),
	}

	return &activityData, nil
}
