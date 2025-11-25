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

	temperatureStats := common.TemperatureStats{}
	var tempSum int
	var tempCount uint

	gpsAccuracyStats := common.GpsAccuracyStats{}
	var gpsSum, gpsCount uint

	altitudeStats := common.AltitudeStats{
		Min: common.NewAltitude(0),
		Max: common.NewAltitude(0),
	}
	var altitudeCount uint

	heartrateStats := common.HeartrateStats{}
	var heartrateSum, heartrateCount uint

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

		var temperature *common.Temperature
		if r.Temperature != basetype.Sint8Invalid {
			temperature = common.NewTemperature(r.Temperature)
			// `Temperature` stats calculation
			// initialize values on first valid `Temperature`
			if tempCount == 0 {
				temperatureStats.Min = temperature
				temperatureStats.Max = temperature
				temperatureStats.Avg = temperature
			}
			// compare min
			if temperature.Value < temperatureStats.Min.Value {
				temperatureStats.Min = temperature
			}
			// compare max
			if temperature.Value > temperatureStats.Max.Value {
				temperatureStats.Max = temperature
			}
			tempCount += 1
			tempSum += int(temperature.Value)
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

		var gpsAccuracy *common.GpsAccuracy
		if r.GpsAccuracy != basetype.Uint8Invalid {
			gpsAccuracy = common.NewGpsAccuracy(r.GpsAccuracy)
			// `GpsAccuracyStats` calculation
			// initialize values on first valid `GpsAccuracy`
			if gpsCount == 0 {
				gpsAccuracyStats.Min = gpsAccuracy
				gpsAccuracyStats.Max = gpsAccuracy
				gpsAccuracyStats.Avg = gpsAccuracy
			}
			// compare min
			if gpsAccuracy.Value < gpsAccuracyStats.Min.Value {
				gpsAccuracyStats.Min = gpsAccuracy
			}
			// compare max
			if gpsAccuracy.Value > gpsAccuracyStats.Max.Value {
				gpsAccuracyStats.Max = gpsAccuracy
			}
			gpsCount += 1
			gpsSum += uint(gpsAccuracy.Value)
		}

		var heartrate *common.Heartrate
		if r.HeartRate != basetype.Uint8Invalid {
			heartrate = common.NewHeartrate(r.HeartRate)
			// `HeartrateStats` calculation
			// initialize values on first valid `Heartrate`
			if heartrateCount == 0 {
				heartrateStats.Min = heartrate
				heartrateStats.Max = heartrate
				heartrateStats.Avg = heartrate
			}
			// compare min
			if heartrate.Value < heartrateStats.Min.Value {
				heartrateStats.Min = heartrate
			}
			// compare max
			if heartrate.Value > heartrateStats.Max.Value {
				heartrateStats.Max = heartrate
			}
			heartrateCount += 1
			heartrateSum += uint(heartrate.Value)
		}

		record := common.RecordData{
			Time:        common.NewTime(r.Timestamp.Local()),
			Distance:    common.NewDistance(distance),
			Speed:       common.NewSpeed(float32(speed)),
			Temperature: temperature,
			Altitude:    common.NewAltitude(altitudeValue),
			GpsAccuracy: gpsAccuracy,
			Heartrate:   heartrate,
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

	// Calculate `Temperature` average
	if tempCount > 0 {
		temperatureStats.Avg = common.NewTemperature(int8(float32(tempSum) / float32(tempCount)))
	}

	// Calculate `GpsAccuracy` average
	if gpsCount > 0 {
		gpsAccuracyStats.Avg = common.NewGpsAccuracy(uint8(float32(gpsSum) / float32(gpsCount)))
	}

	// Calculate `Heartrate` average
	if heartrateCount > 0 {
		heartrateStats.Avg = common.NewHeartrate(uint8(float32(heartrateSum) / float32(heartrateCount)))
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
		Heartrate:     heartrateStats,
	}

	return activityData, nil
}
