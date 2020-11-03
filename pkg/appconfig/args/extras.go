package args

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"os"
	"strings"
)

func Extras(exists func(name string) bool) (extras map[string]string) {
	extras = make(map[string]string)
	args := os.Args[1:]
	var strip []int
	for n := 0; n < len(args); n++ {
		v := args[n]
		if len(v) < 2 {
			continue
		}
		if v == "--" {
			break
		}
		if !strings.HasPrefix(v, "--") {
			continue
		}
		v = v[2:]
		split := strings.SplitN(v, "=", 2)
		if len(split) == 2 {
			key := split[0]
			extras[key] = split[1]
			strip = append(strip, n)
		} else if n == len(args)-1 {
			continue
		} else if strings.HasPrefix(args[n+1], "--") {
			continue
		} else if exists(v) {
			// Flag exists
			n++
		} else {
			key := appconfig.NormalizeKey(v)
			extras[key] = args[n+1]
			strip = append(strip, n, n+1)
			n++
		}
	}

	// Remove arguments we processed
	for n := len(strip) - 1; n >= 0; n-- {
		idx := strip[n]
		args = append(args[:idx], args[idx+1:]...)
	}
	os.Args = append(os.Args[:1], args...)

	return extras
}

