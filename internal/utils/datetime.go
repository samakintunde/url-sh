package utils

import "time"

// Uses RFC 3359 compliant time format
// Example: 1985-04-12T23:20:50.52Z
func ConvertTimeToString(time time.Time) string {
	timeString, err := time.MarshalText()
	if err != nil {
		return ""
	}
	return string(timeString)
}
