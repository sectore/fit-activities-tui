package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sectore/fit-sum-tui/internal/asyncdata"
	"github.com/sectore/fit-sum-tui/internal/common"
)

func GetTotalDistance(data common.ActivityData) uint32 {
	var dist uint32 = 0
	for _, d := range data.Distances {
		dist += d
	}
	return dist
}

func FormatTotalDistance(data common.ActivityData) string {
	var meters = GetTotalDistance(data) / 100 / 1000
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

func FormatTotalTime(data common.ActivityData) string {
	seconds := int(data.TotalTime / 1000)
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

func ActivitiesAreLoading(acts []common.Activity) bool {
	for _, act := range acts {
		if asyncdata.IsLoading(act.Data) {
			return true
		}
	}
	return false
}

func ActivitiesSuccess(acts []common.Activity) int {
	count := 0
	for _, act := range acts {
		if asyncdata.IsSuccess(act.Data) {
			count += 1
		}
	}
	return count
}

func ActivitiesFailures(acts []common.Activity) int {
	count := 0
	for _, act := range acts {
		if asyncdata.IsFailure(act.Data) {
			count += 1
		}
	}
	return count
}
