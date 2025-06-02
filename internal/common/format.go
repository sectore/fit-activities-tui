package common

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	BulletPointBig = "●"
	BulletPoint    = "∙"
)

func FormatLocalTime(time time.Time) string {
	// return data.LocalTime.Format("2006-01-02 15:04")
	return time.Format("02.01.06 15:04")
}

func FormatTotalTime(totalTime uint32) string {
	var s string
	seconds := int(totalTime / 1000)
	if seconds < 60 {
		// Format as ss
		s = strconv.Itoa(seconds)
	} else if seconds < 3600 {
		// Format as mm:ss
		minutes := seconds / 60
		remainingSeconds := seconds % 60
		s = fmt.Sprintf("%02dm %02ds", minutes, remainingSeconds)
	} else if seconds < 86400 {
		// Format as hh:mm:ss
		hours := seconds / 3600
		remainingSeconds := seconds % 3600
		minutes := remainingSeconds / 60
		seconds = remainingSeconds % 60
		s = fmt.Sprintf("%02dh %02dm %02ds", hours, minutes, seconds)
	} else {
		// Format as dd:hh:mm:ss
		days := seconds / 86400
		remainingSeconds := seconds % 86400
		hours := remainingSeconds / 3600
		remainingSeconds = remainingSeconds % 3600
		minutes := remainingSeconds / 60
		seconds = remainingSeconds % 60
		s = fmt.Sprintf("%dd %02dh %02dm %02ds", days, hours, minutes, seconds)
	}
	return s
}

func FormatTotalDistance(dist uint32) string {
	var meters = dist / 100
	if meters >= 1000 {
		km := float64(meters) / 1000
		formatted := fmt.Sprintf("%.1f", km)
		formatted = strings.TrimRight(formatted, "0")
		formatted = strings.TrimRight(formatted, ".")
		return formatted + "km"
	} else {
		return fmt.Sprintf("%dm", meters)
	}
}

func FormatTemperature(t Temperature) string {
	return fmt.Sprintf("%d°C", t)
}

func FormatAscent(v Ascent) string {
	return fmt.Sprintf("%dm", v)
}

func FormatDescent(v Descent) string {
	return fmt.Sprintf("%dm", v)
}
