package fit

import (
	"fmt"
	"math"
	"os"

	"github.com/muktihari/fit/decoder"
	"github.com/muktihari/fit/profile/basetype"
	"github.com/muktihari/fit/profile/filedef"
	"github.com/sectore/fit-activities-tui/internal/common"
)

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

		attitudeValue := r.AltitudeScaled()
		if math.IsNaN(attitudeValue) || math.Float64bits(attitudeValue) == basetype.Float64Invalid {
			attitudeValue = 0
		}

		temperature := r.Temperature
		if temperature == basetype.Sint8Invalid {
			temperature = 0
		}

		distance := r.Distance
		if distance == basetype.Uint32Invalid {
			distance = 0
		}

		speed := r.Speed
		if speed == basetype.Uint16Invalid {
			speed = 0
		}

		gpsAccuracy := r.GpsAccuracy
		if gpsAccuracy == basetype.Uint8Invalid {
			gpsAccuracy = 0
		}

		record := common.RecordData{
			Time:        common.NewTime(r.Timestamp.Local()),
			Distance:    common.NewDistance(distance),
			Speed:       common.NewSpeed(float32(speed)),
			Temperature: common.NewTemperature(float32(temperature)),
			Altitude:    common.NewAltitude(attitudeValue),
			GpsAccuracy: common.NewGpsAccuracy(float32(gpsAccuracy)),
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
	durationStats := common.DurationStats{
		Total:  common.NewDuration(0),
		Active: common.NewDuration(0),
		Pause:  common.NewDuration(0),
	}
	elevationStats := common.ElevationStats{
		Ascents:  common.NewElevation(0),
		Descents: common.NewElevation(0),
	}
	altitudeStats := common.AltitudeStats{
		Min: common.NewAltitude(0),
		Max: common.NewAltitude(0),
	}

	for _, s := range act.Sessions {
		totalDistance.Value += s.TotalDistance

		// Duration stats calculation
		total := s.TotalElapsedTime
		if total != basetype.Uint32Invalid {
			durationStats.Total.Value += total
		}
		active := s.TotalTimerTime
		if active != basetype.Uint32Invalid {
			durationStats.Active.Value += active
		}

		// Elevation stats calculation
		ascentValue := s.TotalAscent
		if ascentValue != basetype.Uint16Invalid {
			elevationStats.Ascents.Value += ascentValue
		}
		descentValue := s.TotalDescent
		if descentValue != basetype.Uint16Invalid {
			elevationStats.Descents.Value += descentValue
		}

		minAltitudeValue := s.MinAltitudeScaled()
		if !math.IsNaN(minAltitudeValue) {
			altitudeStats.Min.Value = minAltitudeValue
		}

		maxAltitudeValue := s.MaxAltitudeScaled()
		if !math.IsNaN(maxAltitudeValue) {
			altitudeStats.Max.Value = maxAltitudeValue
		}
	}
	// calculate pause
	durationStats.Pause.Value = durationStats.Total.Value - durationStats.Active.Value

	var activityData = &common.ActivityData{
		Duration:      durationStats,
		TotalDistance: totalDistance,
		Temperature:   temperatureStats,
		Speed:         speedStats,
		Elevation:     elevationStats,
		NoSessions:    uint32(noSessions),
		Records:       records,
		GpsAccuracy:   gpsAccuracyStats,
		Altitude:      altitudeStats,
	}

	return activityData, nil
}
