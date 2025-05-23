package common

import (
	"time"

	"github.com/sectore/fit-sum-tui/internal/asyncdata"
)

type ActivityData struct {
	LocalTime time.Time
	TotalTime uint32
	Distances []uint32
}

type ActivityAD = asyncdata.AsyncData[error, ActivityData]

func ActivityLoading(prevData *ActivityData) ActivityAD {
	return asyncdata.Loading[error, ActivityData](prevData)
}

func ActivityFailure(err error) ActivityAD {
	return asyncdata.Failure[error, ActivityData](err)
}

func ActivitySuccess(data ActivityData) ActivityAD {
	return asyncdata.Success[error, ActivityData](data)
}

type Activity struct {
	Path string
	Data ActivityAD
}
