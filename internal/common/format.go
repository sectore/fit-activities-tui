package common

import (
	"time"
)

func FormatLocalTime(time time.Time) string {
	// return data.LocalTime.Format("2006-01-02 15:04")
	return time.Format("02.01.06 15:04")
}
