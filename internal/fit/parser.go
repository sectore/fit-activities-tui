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

	altitudeStats := common.AltitudeStats{
		Min: common.NewAltitude(0),
		Max: common.NewAltitude(0),
	}
	var altitudeCount uint

	for _, r := range act.Records {

		// Use `EnhancedAltitude` if available, otherwise fallback to `Altitude`
		// This ensures compatibility (Garmin vs. Wahoo)
		altitudeValue := r.EnhancedAltitudeScaled()
		if math.Float64bits(altitudeValue) == basetype.Float64Invalid {
			altitudeValue = r.AltitudeScaled()
			if math.Float64bits(altitudeValue) == basetype.Float64Invalid {
				altitudeValue = 0
			}
		}

		temperature := r.Temperature
		if temperature == basetype.Sint8Invalid {
			temperature = 0
		}

		distance := r.Distance
		if distance == basetype.Uint32Invalid {
			distance = 0
		}

		// Use `EnhancedSpeed` if available, otherwise fallback to `Speed`
		// This ensures compatibility (Garmin: `EnhancedSpeed`, Wahoo: both)
		speed := r.EnhancedSpeed
		if speed == basetype.Uint32Invalid {
			speed = uint32(r.Speed)
			if r.Speed == basetype.Uint16Invalid {
				speed = 0
			}
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
			Altitude:    common.NewAltitude(altitudeValue),
			GpsAccuracy: common.NewGpsAccuracy(float32(gpsAccuracy)),
		}
		records = append(records, record)

		// Speed stats calculation
		speed_f := float32(speed)
		if speedStats.Max.Value < speed_f {
			speedStats.Max.Value = speed_f
		}
		speedCount += 1
		speedTotal += uint(speed)

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

		// Altitude stats calculation
		// Note: Since `altitudeValue` is already validated above
		// we don't check for invalid values here
		if altitudeCount == 0 {
			altitudeStats.Min.Value = altitudeValue
			altitudeStats.Max.Value = altitudeValue
		}
		if altitudeValue < altitudeStats.Min.Value {
			altitudeStats.Min.Value = altitudeValue
		}
		if altitudeValue > altitudeStats.Max.Value {
			altitudeStats.Max.Value = altitudeValue
		}
		altitudeCount += 1
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
