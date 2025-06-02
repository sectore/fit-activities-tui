package common

import (
	"fmt"
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
