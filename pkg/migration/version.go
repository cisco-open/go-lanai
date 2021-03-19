package migration

import (
	"database/sql/driver"
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

type Version []int

func (v Version) Lt(other Version) bool {
	maxLen := len(v)
	if len(other) < maxLen {
		maxLen = len(other)
	}

	for n := 0; n < maxLen; n++ {
		if v[n] < other[n] {
			return true
		}
		if v[n] > other[n] {
			return false
		}
	}

	return len(v) < len(other)
}

func (v Version) String() string {
	var sb = strings.Builder{}
	for _, v := range v {
		if sb.Len() > 0 {
			sb.WriteRune('.')
		}
		sb.WriteString(strconv.Itoa(v))
	}
	return sb.String()
}

func (v Version) Equals(o Version) bool {
	if len(v) != len(o) {
		return false
	}

	for i, n := range v {
		if n != o[i] {
			return false
		}
	}

	return true
}

func (v *Version) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return v.scanString(string(src))
	case string:
		return v.scanString(src)
	case nil:
		*v = nil
		return nil
	}
	return fmt.Errorf("pq: cannot convert %T to StringArray", src)
}

func (v Version) Value() (driver.Value, error) {
	if v == nil {
		return nil, nil
	}
	return v.String(), nil
}

func (v Version) GormDataType() string {
	return "string"
}

func (v *Version) scanString(src string) error {
	result, err := fromString(src)
	*v = result
	return err
}

func fromString(source string) (Version, error) {
	parts := strings.Split(source, ".")
	var numbers []int

	if len(parts) == 0 {
		return Version{}, errors.New("Version must have at least one numeric component")
	}

	for _, part := range parts {
		if number, err := strconv.Atoi(part); err != nil {
			return Version{}, errors.Wrap(err, "Cannot parse component as integer")
		} else {
			numbers = append(numbers, number)
		}
	}

	return numbers, nil
}