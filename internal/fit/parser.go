package fit

import (
	"errors"
	"math"
	"os"

	"github.com/muktihari/fit/decoder"
	"github.com/muktihari/fit/profile/basetype"
	"github.com/muktihari/fit/profile/filedef"
	"github.com/muktihari/fit/profile/mesgdef"
	"github.com/sectore/fit-activities-tui/internal/common"
)

func parseSpeed(records []*mesgdef.Record) common.SpeedStats {
	var list []uint16
	for _, rs := range records {
		value := rs.Speed
		// don't count invalid values
		if value != basetype.Uint16Invalid {
			list = append(list, value)
		}
	}

	l := len(list)

	if l == 0 {
		return common.SpeedStats{
			Avg: common.NewSpeed(0),
			Max: common.NewSpeed(0),
		}
	}

	var total uint64
	var max uint16
	for _, value := range list {
		total += uint64(value)
		if max < value {
			max = value
		}
	}
	avg := math.Round(float64(total) / float64(l))
	return common.SpeedStats{
		Avg: common.NewSpeed(uint16(avg)),
		Max: common.NewSpeed(max),
	}

}

func parseGpsAccurancies(records []*mesgdef.Record) common.GpsAccuracyStat {

	var list []uint8
	for _, record := range records {
		value := record.GpsAccuracy
		// don't count invalid values
		if value != basetype.Uint8Invalid {
			list = append(list, value)
		}
	}
	min := common.NewGpsAccuracy(0)
	if len(list) > 0 {
		min = common.NewGpsAccuracy(list[0])
	}
	gpsAccurancy := common.GpsAccuracyStat{
		Avg: common.NewGpsAccuracy(0),
		Min: min,
		Max: common.NewGpsAccuracy(0),
	}
	total := int(0)
	for _, value := range list {
		total += int(value)
		if value < gpsAccurancy.Min.Value {
			gpsAccurancy.Min.Value = value
		}
		if value > gpsAccurancy.Max.Value {
			gpsAccurancy.Max.Value = value
		}

	}
	gpsAccurancy.Avg.Value = uint8(total / len(list))

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

	var ad = common.ActivityData{
		LocalTime:      act.Activity.LocalTimestamp,
		TotalTime:      act.Activity.TotalTimerTime,
		TotalDistances: distances,
		Temperatures:   temperatures,
		Speed:          parseSpeed(act.Records),
		Descents:       descents,
		Ascents:        ascents,
		NoSessions:     uint32(noSessions),
		NoRecords:      uint32(noRecords),
		GpsAccuracy:    parseGpsAccurancies(act.Records),
	}

	return &ad, nil
}
