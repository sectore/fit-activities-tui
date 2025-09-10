package fit

import (
	"fmt"
	"os"

	"github.com/muktihari/fit/decoder"
	"github.com/muktihari/fit/profile/basetype"
	"github.com/muktihari/fit/profile/filedef"
	"github.com/muktihari/fit/profile/mesgdef"
	"github.com/sectore/fit-activities-tui/internal/common"
)

func parseDuration(ss []*mesgdef.Session) common.DurationStats {
	time := common.DurationStats{
		Total:  common.NewDuration(0),
		Active: common.NewDuration(0),
		Pause:  common.NewDuration(0),
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
		return nil, fmt.Errorf("no Activity found (file %s)", file)
	}

	noSessions := len(act.Sessions)
	if noSessions <= 0 {
		return nil, fmt.Errorf("no Sessions found (file %s)", file)
	}

	if len(act.Records) <= 0 {
		return nil, fmt.Errorf("no Records found (file %s)", file)
	}

	var records []common.RecordData
	speedStats := common.SpeedStats{
		Avg: common.NewSpeed(0),
		Max: common.NewSpeed(0),
	}
	var speedCount, speedTotal uint

	temperatureStats := common.TemperatureStats{
		Min: common.NewTemperature(0),
		Max: common.NewTemperature(0),
		Avg: common.NewTemperature(0),
	}
	var tempSum, tempCount uint

	gpsAccuracyStats := common.GpsAccuracyStats{
		Avg: common.NewGpsAccuracy(0),
		Min: common.NewGpsAccuracy(0),
		Max: common.NewGpsAccuracy(0),
	}
	var gpsSum, gpsCount uint

	for _, r := range act.Records {
		record := common.RecordData{
			Time:        common.NewTime(r.Timestamp.Local()),
			Distance:    common.NewDistance(r.Distance),
			Speed:       common.NewSpeed(float32(r.Speed)),
			Temperature: common.NewTemperature(float32(r.Temperature)),
			Elevation:   common.NewElevation(r.Altitude),
			GpsAccuracy: common.NewGpsAccuracy(float32(r.GpsAccuracy)),
		}
		records = append(records, record)

		// Speed stats calculation
		speedValue := r.Speed
		if speedValue != basetype.Uint16Invalid {
			speedValue_f := float32(r.Speed)
			if speedStats.Max.Value < speedValue_f {
				speedStats.Max.Value = speedValue_f
			}
			speedCount += 1
			speedTotal += uint(speedValue)
		}

		// Calculate speed average
		if speedCount > 0 {
			avg := float32(speedTotal / speedCount)
			speedStats.Avg.Value = avg
		}

		// Temperature stats calculation
		tempValue := r.Temperature
		if tempValue != basetype.Sint8Invalid {
			tempValue_f := float32(tempValue)
			// override default zero value
			if tempCount == 0 {
				temperatureStats.Min.Value = tempValue_f
			}
			// compare min
			if tempValue_f < temperatureStats.Min.Value {
				temperatureStats.Min.Value = tempValue_f
			}
			// compare max
			if tempValue_f > temperatureStats.Max.Value {
				temperatureStats.Max.Value = tempValue_f
			}
			tempCount += 1
			tempSum += uint(tempValue)
		}

		// Calculate temperature average
		if tempCount > 0 {
			temperatureStats.Avg.Value = float32(tempSum / tempCount)
		}

		// GPS accuracy stats calculation
		gpsValue := r.GpsAccuracy
		if gpsValue != basetype.Uint8Invalid {
			gpsValue_f := float32(gpsValue)
			// override default zero value
			if gpsCount == 0 {
				gpsAccuracyStats.Min.Value = gpsValue_f
			}
			// compare min
			if gpsValue_f < gpsAccuracyStats.Min.Value {
				gpsAccuracyStats.Min.Value = gpsValue_f
			}
			// compare max
			if gpsValue_f > gpsAccuracyStats.Max.Value {
				gpsAccuracyStats.Max.Value = gpsValue_f
			}
			gpsCount += 1
			gpsSum += uint(gpsValue)
		}
	}

	// Calculate GPS accuracy average
	if gpsCount > 0 {
		gpsAccuracyStats.Avg.Value = float32(gpsSum / gpsCount)
	}

	totalDistance := common.NewDistance(0)
	for _, s := range act.Sessions {
		totalDistance.Value += s.TotalDistance
	}

	var activityData = common.ActivityData{
		Duration:      parseDuration(act.Sessions),
		TotalDistance: totalDistance,
		Temperature:   temperatureStats,
		Speed:         speedStats,
		Elevation:     parseElevation(act.Sessions),
		NoSessions:    uint32(noSessions),
		Records:       records,
		GpsAccuracy:   gpsAccuracyStats,
	}

	return &activityData, nil
}
