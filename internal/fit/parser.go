package fit

import (
	"errors"
	"os"

	"github.com/muktihari/fit/decoder"
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

	speeds := make([]common.Speed, noRecords)
	for i, rs := range act.Records {
		speeds[i] = rs.Speed / 1000 // Record.Speed = Scale: 1000
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
	}

	return &ad, nil
}
