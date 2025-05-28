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

	distances := make([]uint32, len(act.Sessions))
	for i, sess := range act.Sessions {
		distances[i] = sess.TotalDistance
	}

	var ad = common.ActivityData{
		LocalTime:      act.Activity.LocalTimestamp,
		TotalTime:      act.Activity.TotalTimerTime,
		TotalDistances: distances,
	}

	return &ad, nil
}
