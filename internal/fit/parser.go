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
	speedStats := common.SpeedStats{}
	var speedCount, speedTotal uint

	temperatureStats := common.TemperatureStats{}
	var tempSum int
	var tempCount uint

	gpsAccuracyStats := common.GpsAccuracyStats{}
	var gpsSum, gpsCount uint

	altitudeStats := common.AltitudeStats{}
	var altitudeCount uint

	heartrateStats := common.HeartrateStats{}
	var heartrateSum, heartrateCount uint

	for _, r := range act.Records {

		// Use `EnhancedAltitudeScaled` if available, otherwise fallback to `AltitudeScaled`
		// This ensures compatibility (Garmin vs. Wahoo)
		var altitude *common.Altitude
		if math.Float64bits(r.EnhancedAltitudeScaled()) != basetype.Float64Invalid {
			altitude = common.NewAltitude(r.EnhancedAltitudeScaled())
		} else if math.Float64bits(r.AltitudeScaled()) != basetype.Float64Invalid {
			altitude = common.NewAltitude(r.AltitudeScaled())
		}

		if altitude != nil {
			// `AltitudeStats` calculation
			// initialize `max`/`min` on first valid `Altitude`
			if altitudeCount == 0 {
				altitudeStats.Min = altitude
				altitudeStats.Max = altitude
			}

			if altitude.Value < altitudeStats.Min.Value {
				altitudeStats.Min = altitude
			}
			if altitude.Value > altitudeStats.Max.Value {
				altitudeStats.Max = altitude
			}

			altitudeCount += 1
		}

		var temperature *common.Temperature
		if r.Temperature != basetype.Sint8Invalid {
			temperature = common.NewTemperature(r.Temperature)
			// `Temperature` stats calculation
			// initialize min/max on first valid `Temperature`
			if tempCount == 0 {
				temperatureStats.Min = temperature
				temperatureStats.Max = temperature
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

		var speed *common.Speed
		// Use `EnhancedSpeed` if available, otherwise fallback to `Speed`
		// This ensures compatibility (Garmin: `EnhancedSpeed`, Wahoo: both)
		if r.EnhancedSpeed != basetype.Uint32Invalid {
			speed = common.NewSpeed(float32(r.EnhancedSpeed))
		} else if r.Speed != basetype.Uint16Invalid {
			speed = common.NewSpeed(float32(r.Speed))
		}

		if speed != nil {
			// `SpeedStats` calculation
			// initialize `max` on first valid `Speed`
			if speedCount == 0 {
				speedStats.Max = speed
			}

			if speedStats.Max.Value < speed.Value {
				speedStats.Max = speed
			}
			speedCount += 1
			speedTotal += uint(speed.Value)
		}

		var gpsAccuracy *common.GpsAccuracy
		if r.GpsAccuracy != basetype.Uint8Invalid {
			gpsAccuracy = common.NewGpsAccuracy(r.GpsAccuracy)
			// `GpsAccuracyStats` calculation
			// initialize min/max on first valid `GpsAccuracy`
			if gpsCount == 0 {
				gpsAccuracyStats.Min = gpsAccuracy
				gpsAccuracyStats.Max = gpsAccuracy
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
			// initialize min/max on first valid `Heartrate`
			if heartrateCount == 0 {
				heartrateStats.Min = heartrate
				heartrateStats.Max = heartrate
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
			Speed:       speed,
			Temperature: temperature,
			Altitude:    altitude,
			GpsAccuracy: gpsAccuracy,
			Heartrate:   heartrate,
		}
		records = append(records, record)

	}

	// Calculate `Speed` average
	if speedCount > 0 {
		speedStats.Avg = common.NewSpeed(float32(speedTotal / speedCount))
	}

	// Calculate `Temperature` average
	// Use `math.Round` for symmetric handling: truncation has sign-dependent bias
	// (positive: 16.5°C → 16°C; negative: -0.5°C → 0°C truncates toward zero)
	if tempCount > 0 {
		temperatureStats.Avg = common.NewTemperature(int8(math.Round(float64(tempSum) / float64(tempCount))))
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
	durationStats := common.DurationStats{}
	elevationStats := common.ElevationStats{}

	for _, s := range act.Sessions {
		totalDistance.Value += s.TotalDistance

		// Duration stats calculation
		if s.TotalElapsedTime != basetype.Uint32Invalid {
			value := s.TotalElapsedTime
			if durationStats.Total == nil {
				d := common.NewDuration(value)
				durationStats.Total = &d
			} else {
				durationStats.Total.Value += value
			}
		}
		if s.TotalTimerTime != basetype.Uint32Invalid {
			value := s.TotalTimerTime
			if durationStats.Active == nil {
				d := common.NewDuration(value)
				durationStats.Active = &d
			} else {
				durationStats.Active.Value += value
			}
		}

		// Elevation stats calculation
		if s.TotalAscent != basetype.Uint16Invalid {
			value := s.TotalAscent
			if elevationStats.Ascents == nil {
				elevationStats.Ascents = common.NewElevation(value)
			} else {
				elevationStats.Ascents.Value += value
			}
		}

		if s.TotalDescent != basetype.Uint16Invalid {
			value := s.TotalAscent
			if elevationStats.Descents == nil {
				elevationStats.Descents = common.NewElevation(value)
			} else {
				elevationStats.Descents.Value += value
			}
		}
	}

	// calculate pause
	if durationStats.Total != nil && durationStats.Active != nil {
		value := durationStats.Total.Value - durationStats.Active.Value
		d := common.NewDuration(value)
		durationStats.Pause = &d
	}

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
