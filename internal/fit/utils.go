package fit

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/muktihari/fit/profile/filedef"
)

func GetLocalTime(act *filedef.Activity) time.Time {
	return act.Activity.Timestamp
}

func GetTotalDistance(act *filedef.Activity) uint32 {
	var dist uint32 = 0
	for _, rec := range act.Records {
		dist += rec.Distance
	}
	return dist
}

func FormatTotalDistance(act *filedef.Activity) string {
	var meters = GetTotalDistance(act) / 100 / 1000
	if meters >= 1000 {
		km := float64(meters) / 1000
		formatted := fmt.Sprintf("%.3f", km)
		formatted = strings.TrimRight(formatted, "0")
		formatted = strings.TrimRight(formatted, ".")
		return formatted + "km"
	} else {
		return fmt.Sprintf("%dm", meters)
	}

}

func GetTotalTime(act *filedef.Activity) uint32 {
	return act.Activity.TotalTimerTime
}

func FormatTotalTime(act *filedef.Activity) string {
	seconds := int(GetTotalTime(act) / 1000)
	if seconds < 60 {
		// Format as ss
		return strconv.Itoa(seconds)
	} else if seconds < 3600 {
		// Format as mm:ss
		minutes := seconds / 60
		remainingSeconds := seconds % 60
		return fmt.Sprintf("%02d:%02d", minutes, remainingSeconds)
	} else if seconds < 86400 {
		// Format as hh:mm:ss
		hours := seconds / 3600
		remainingSeconds := seconds % 3600
		minutes := remainingSeconds / 60
		seconds = remainingSeconds % 60
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	} else {
		// Format as dd:hh:mm:ss
		days := seconds / 86400
		remainingSeconds := seconds % 86400
		hours := remainingSeconds / 3600
		remainingSeconds = remainingSeconds % 3600
		minutes := remainingSeconds / 60
		seconds = remainingSeconds % 60
		return fmt.Sprintf("%d:%02d:%02d:%02d", days, hours, minutes, seconds)
	}
}
