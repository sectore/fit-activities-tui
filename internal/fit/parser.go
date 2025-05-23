package fit

import (
	"errors"
	"os"

	"github.com/muktihari/fit/decoder"
	"github.com/muktihari/fit/profile/filedef"
)

type Parser struct {
	files []string
}

func New(files []string) Parser {
	return Parser{files}
}

func ParseFile(file string) (*filedef.Activity, error) {
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

	activity, ok := lis.File().(*filedef.Activity)
	if !ok {
		return nil, errors.New("expected an Activity")
	}
	// fmt.Printf("Distance: %.2f km\n", activity.Sessions[0].TotalDistanceScaled()/1000)

	return activity, nil
}
