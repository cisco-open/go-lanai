package tenant_hierarchy_accessor

import (
	"errors"
	"fmt"
	"strings"
)

const spoPrefix = "spo"
func BuildSpsString(subject string, predict string, object... string) string {
	if len(object) == 0 {
		return fmt.Sprintf("%s:%s:%s", spoPrefix, subject, predict)
	} else {
		return fmt.Sprintf("%s:%s:%s:%s", spoPrefix, subject, predict, object[0])
	}
}

func GetObjectOfSpo(spo string) (string, error) {
	parts := strings.Split(spo, ":")

	if len(parts) == 4 {
		return parts[3], nil
	} else {
		return "", errors.New("spo relation has no object part")
	}
}

func ZInclusive(min string) string {
	return fmt.Sprintf("[%s", min)
}