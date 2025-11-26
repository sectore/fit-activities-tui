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
		var altitudePtr *common.Altitude
		if math.Float64bits(r.EnhancedAltitudeScaled()) != basetype.Float64Invalid {
			altitude := common.NewAltitude(r.EnhancedAltitudeScaled())
			altitudePtr = common.Ptr(altitude)
		} else if math.Float64bits(r.AltitudeScaled()) != basetype.Float64Invalid {
			altitude := common.NewAltitude(r.AltitudeScaled())
			altitudePtr = common.Ptr(altitude)
		}

		if altitudePtr != nil {
			// `AltitudeStats` calculation
			// initialize `max`/`min` on first valid `Altitude`
			if altitudeCount == 0 {
				altitudeStats.Min = altitudePtr
				altitudeStats.Max = altitudePtr
			}

			if altitudePtr.Value < altitudeStats.Min.Value {
				altitudeStats.Min = altitudePtr
			}
			if altitudePtr.Value > altitudeStats.Max.Value {
				altitudeStats.Max = altitudePtr
			}

			altitudeCount += 1
		}

		var temperaturePtr *common.Temperature
		if r.Temperature != basetype.Sint8Invalid {
			temperature := common.NewTemperature(r.Temperature)
			temperaturePtr = common.Ptr(temperature)
			// `Temperature` stats calculation
			// initialize min/max on first valid `Temperature`
			if tempCount == 0 {
				temperatureStats.Min = temperaturePtr
				temperatureStats.Max = temperaturePtr
			}
			// compare min
			if temperature.Value < temperatureStats.Min.Value {
				temperatureStats.Min = temperaturePtr
			}
			// compare max
			if temperature.Value > temperatureStats.Max.Value {
				temperatureStats.Max = temperaturePtr
			}
			tempCount += 1
			tempSum += int(temperature.Value)
		}

		var distancePtr *common.Distance
		if r.Distance != basetype.Uint32Invalid {
			distance := common.NewDistance(r.Distance)
			distancePtr = common.Ptr(distance)
		}

		var speedPtr *common.Speed
		// Use `EnhancedSpeed` if available, otherwise fallback to `Speed`
		// This ensures compatibility (Garmin: `EnhancedSpeed`, Wahoo: both)
		if r.EnhancedSpeed != basetype.Uint32Invalid {
			speed := common.NewSpeed(float32(r.EnhancedSpeed))
			speedPtr = common.Ptr(speed)
		} else if r.Speed != basetype.Uint16Invalid {
			speed := common.NewSpeed(float32(r.Speed))
			speedPtr = common.Ptr(speed)
		}

		if speedPtr != nil {
			// `SpeedStats` calculation
			// initialize `max` on first valid `Speed`
			if speedCount == 0 {
				speedStats.Max = speedPtr
			}

			if speedStats.Max.Value < speedPtr.Value {
				speedStats.Max = speedPtr
			}
			speedCount += 1
			speedTotal += uint(speedPtr.Value)
		}

		var gpsAccuracyPtr *common.GpsAccuracy
		if r.GpsAccuracy != basetype.Uint8Invalid {
			gpsAccuracy := common.NewGpsAccuracy(r.GpsAccuracy)
			gpsAccuracyPtr = common.Ptr(gpsAccuracy)
			// `GpsAccuracyStats` calculation
			// initialize min/max on first valid `GpsAccuracy`
			if gpsCount == 0 {
				gpsAccuracyStats.Min = gpsAccuracyPtr
				gpsAccuracyStats.Max = gpsAccuracyPtr
			}
			// compare min
			if gpsAccuracy.Value < gpsAccuracyStats.Min.Value {
				gpsAccuracyStats.Min = gpsAccuracyPtr
			}
			// compare max
			if gpsAccuracy.Value > gpsAccuracyStats.Max.Value {
				gpsAccuracyStats.Max = gpsAccuracyPtr
			}
			gpsCount += 1
			gpsSum += uint(gpsAccuracy.Value)
		}

		var heartratePtr *common.Heartrate
		if r.HeartRate != basetype.Uint8Invalid {
			heartrate := common.NewHeartrate(r.HeartRate)
			heartratePtr = common.Ptr(heartrate)
			// `HeartrateStats` calculation
			// initialize min/max on first valid `Heartrate`
			if heartrateCount == 0 {
				heartrateStats.Min = heartratePtr
				heartrateStats.Max = heartratePtr
			}
			// compare min
			if heartrate.Value < heartrateStats.Min.Value {
				heartrateStats.Min = heartratePtr
			}
			// compare max
			if heartrate.Value > heartrateStats.Max.Value {
				heartrateStats.Max = heartratePtr
			}
			heartrateCount += 1
			heartrateSum += uint(heartrate.Value)
		}

		record := common.RecordData{
			Time:        common.NewTime(r.Timestamp.Local()),
			Distance:    distancePtr,
			Speed:       speedPtr,
			Temperature: temperaturePtr,
			Altitude:    altitudePtr,
			GpsAccuracy: gpsAccuracyPtr,
			Heartrate:   heartratePtr,
		}
		records = append(records, record)

	}

	// Calculate `Speed` average
	if speedCount > 0 {
		speed := common.NewSpeed(float32(speedTotal / speedCount))
		speedStats.Avg = common.Ptr(speed)
	}

	// Calculate `Temperature` average
	// Use `math.Round` for symmetric handling: truncation has sign-dependent bias
	// (positive: 16.5°C → 16°C; negative: -0.5°C → 0°C truncates toward zero)
	if tempCount > 0 {
		temperature := common.NewTemperature(int8(math.Round(float64(tempSum) / float64(tempCount))))
		temperatureStats.Avg = common.Ptr(temperature)
	}

	// Calculate `GpsAccuracy` average
	if gpsCount > 0 {
		gpsAccuracy := common.NewGpsAccuracy(uint8(float32(gpsSum) / float32(gpsCount)))
		gpsAccuracyStats.Avg = common.Ptr(gpsAccuracy)
	}

	// Calculate `Heartrate` average
	if heartrateCount > 0 {
		heartrate := common.NewHeartrate(uint8(float32(heartrateSum) / float32(heartrateCount)))
		heartrateStats.Avg = common.Ptr(heartrate)
	}

	var totalDistance *common.Distance
	durationStats := common.DurationStats{}
	elevationStats := common.ElevationStats{}

	for _, s := range act.Sessions {

		if s.TotalDistance != basetype.Uint32Invalid {
			value := s.TotalDistance
			if totalDistance == nil {
				d := common.NewDistance(value)
				totalDistance = common.Ptr(d)
			} else {
				// update inner `Value`
				totalDistance.Value += value
			}
		}

		// Duration stats calculation
		if s.TotalElapsedTime != basetype.Uint32Invalid {
			value := s.TotalElapsedTime
			if durationStats.Total == nil {
				d := common.NewDuration(value)
				durationStats.Total = common.Ptr(d)
			} else {
				// update inner `Value`
				durationStats.Total.Value += value
			}
		}
		if s.TotalTimerTime != basetype.Uint32Invalid {
			value := s.TotalTimerTime
			if durationStats.Active == nil {
				d := common.NewDuration(value)
				durationStats.Active = common.Ptr(d)
			} else {
				// update inner `Value`
				durationStats.Active.Value += value
			}
		}

		// Elevation stats calculation
		if s.TotalAscent != basetype.Uint16Invalid {
			value := s.TotalAscent
			if elevationStats.Ascents == nil {
				e := common.NewElevation(value)
				elevationStats.Ascents = common.Ptr(e)
			} else {
				// update inner `Value`
				elevationStats.Ascents.Value += value
			}
		}

		if s.TotalDescent != basetype.Uint16Invalid {
			value := s.TotalDescent
			if elevationStats.Descents == nil {
				e := common.NewElevation(value)
				elevationStats.Descents = common.Ptr(e)
			} else {
				// update inner `Value`
				elevationStats.Descents.Value += value
			}
		}
	}

	// calculate pause
	if durationStats.Total != nil && durationStats.Active != nil {
		d := common.NewDuration(durationStats.Total.Value - durationStats.Active.Value)
		durationStats.Pause = common.Ptr(d)
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
