package utils

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	ISO8601Seconds      = "2006-01-02T15:04:05Z07:00" //time.RFC3339
	ISO8601Milliseconds = "2006-01-02T15:04:05.000Z07:00"
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

type Duration time.Duration

// json.MarshalJSON
func (d *Duration) MarshalJSON() ([]byte, error) {
	if d == nil {
		return []byte(`"0s"`), nil
	}
	str := fmt.Sprintf(`"%s"`, time.Duration(*d).String())
	return []byte(str), nil
}

// json.UnmarshalJSON
func (d *Duration) UnmarshalJSON(data []byte) error {
	if d == nil {
		return errors.New("duration poiter is nil")
	}

	str := strings.Trim(string(data), `"`)
	parsed, e := time.ParseDuration(str)
	if e != nil {
		return e
	}

	*d = Duration(parsed)
	return nil
}