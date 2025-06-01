package fit

import (
	"errors"
	"os"

	"github.com/muktihari/fit/decoder"
	"github.com/muktihari/fit/profile/filedef"
	"github.com/muktihari/fit/profile/mesgdef"
	"github.com/sectore/fit-activities-tui/internal/common"
)

func parseGpsAccurancies(records []*mesgdef.Record) common.GpsAccuracyStat {

	var gpsAccurancies []uint8
	for _, record := range records {
		value := record.GpsAccuracy
		// ignore overflow values
		if value < 255 {
			gpsAccurancies = append(gpsAccurancies, value)
		}
	}
	zero := common.GpsAccuracy{Value: 0}
	min := zero
	if len(gpsAccurancies) > 0 {
		min = common.GpsAccuracy{Value: gpsAccurancies[0]}
	}
	gpsAccurancy := common.GpsAccuracyStat{
		Avg: zero,
		Min: min,
		Max: zero,
	}
	gpsAccurancyTotal := int(0)
	for _, value := range gpsAccurancies {
		gpsAccurancyTotal += int(value)
		if value < gpsAccurancy.Min.Value {
			gpsAccurancy.Min.Value = value
		}
		if value > gpsAccurancy.Max.Value {
			gpsAccurancy.Max.Value = value
		}

	}
	gpsAccurancy.Avg.Value = uint8(gpsAccurancyTotal / len(gpsAccurancies))

	return gpsAccurancy
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

	distances := make([]uint32, noSessions)
	for i, ss := range act.Sessions {
		distances[i] = ss.TotalDistance
	}

	ascents := make([]uint16, noSessions)
	for i, ss := range act.Sessions {
		ascents[i] = ss.TotalAscent
	}

	descents := make([]uint16, noSessions)
	for i, ss := range act.Sessions {
		descents[i] = ss.TotalDescent
	}

	temperatures := make([]common.Temperature, noRecords)
	for i, rs := range act.Records {
		temperatures[i] = rs.Temperature
	}

	var speeds []common.Speed
	for _, rs := range act.Records {
		// don't count overflows
		if rs.Speed < 65535 {
			speeds = append(speeds, rs.Speed/1000) // Record.Speed = Scale: 1000
		}
	}

	var ad = common.ActivityData{
		LocalTime:      act.Activity.LocalTimestamp,
		TotalTime:      act.Activity.TotalTimerTime,
		TotalDistances: distances,
		Temperatures:   temperatures,
		Speeds:         speeds,
		Descents:       descents,
		Ascents:        ascents,
		NoSessions:     uint32(noSessions),
		NoRecords:      uint32(noRecords),
		GpsAccuracy:    parseGpsAccurancies(act.Records),
	}

	return &ad, nil
}
