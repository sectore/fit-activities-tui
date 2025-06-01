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

func parseSpeed(records []*mesgdef.Record) common.SpeedStats {
	var list []uint16
	for _, rs := range records {
		value := rs.Speed
		// don't count invalid values
		if value != basetype.Uint16Invalid {
			list = append(list, uint16(value/1000)) // Record.Speed = Scale: 1000
		}
	}

	l := len(list)

	if l == 0 {
		return common.SpeedStats{
			Avg: common.NewSpeed(0),
			Max: common.NewSpeed(0),
		}
	}

	var total, max uint16
	for _, value := range list {
		total += value
		if max < value {
			max = value
		}
	}
	avg := int(total) / l
	return common.SpeedStats{
		Avg: common.NewSpeed(uint16(avg)),
		Max: common.NewSpeed(max),
	}

}

func parseSpeed2(ss []*mesgdef.Session) common.SpeedStats {
	var avgs []uint16
	var maxs []uint16
	for _, s := range ss {
		avg := s.AvgSpeed
		// don't count invalid values
		if avg != basetype.Uint16Invalid {
			avgs = append(avgs, avg/1000)
		}
		max := s.MaxSpeed
		// don't count invalid values
		if max != basetype.Uint16Invalid {
			maxs = append(maxs, max/1000)
		}
	}

	avg := int(0)
	for _, value := range avgs {
		avg += int(value)
	}
	avg = avg / len(maxs)

	max := int(0)
	for _, value := range maxs {
		max += int(value)
	}
	max = max / len(maxs)

	return common.SpeedStats{
		Max: common.NewSpeed(uint16(max)),
		Avg: common.NewSpeed(uint16(avg)),
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
		// Speed:          parseSpeed2(act.Sessions),
		Descents:    descents,
		Ascents:     ascents,
		NoSessions:  uint32(noSessions),
		NoRecords:   uint32(noRecords),
		GpsAccuracy: parseGpsAccurancies(act.Records),
	}

	return &ad, nil
}
