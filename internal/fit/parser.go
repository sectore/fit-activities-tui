package fit

import (
	"errors"
	"os"

	"github.com/muktihari/fit/decoder"
	"github.com/muktihari/fit/profile/filedef"
	"github.com/sectore/fit-sum-tui/internal/common"
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

	distances := make([]uint32, len(act.Records))
	for i, rec := range act.Records {
		distances[i] = rec.Distance
	}

	var ad = common.ActivityData{
		LocalTime: act.Activity.Timestamp,
		TotalTime: act.Activity.TotalTimerTime,
		Distances: distances,
	}

	return &ad, nil
}
