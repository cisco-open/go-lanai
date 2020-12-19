package utils

import "time"

const (
	ISO8601Seconds      = "2006-01-02T15:04:05Z"
	ISO8601Milliseconds = "2006-01-02T15:04:05.000Z"
)

func ParseTimeISO8601(v string) time.Time {
	parsed, err := time.Parse(ISO8601Seconds, v)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func ParseTime(layout, v string) time.Time {
	parsed, err := time.Parse(layout, v)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func ParseDuration(v string) time.Duration {
	parsed, err := time.ParseDuration(v)
	if err != nil {
		return time.Duration(0)
	}
	return parsed
}
