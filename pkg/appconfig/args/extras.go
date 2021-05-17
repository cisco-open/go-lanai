package args

import (
	"os"
	"strings"
)

// ExtraFlags parse original CLI flags (before standalone "--") and accepts both --flag=value and --flag value format.
// This method is used to parse the flags not pre-defined by our application. (i.e. flags like --help, --profile)
func ExtraFlags(skip func(name string) bool) (extras map[string]string) {
	extras = make(map[string]string)
	args := os.Args[1:]
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
		if len(split) == 2 && !skip(split[0]){
			key := split[0]
			extras[key] = split[1]
		} else if n == len(args)-1 {
			continue
		} else if strings.HasPrefix(args[n+1], "--") {
			continue
		} else if skip(v) {
			// skip this flag. we do n++ since if we ended up here, we are expecting the next argument to be the value
			n++
		} else {
			key := v
			extras[key] = args[n+1]
			n++
		}
	}

	return extras
}

// ExtraKVArgs parse original CLI arguments (after standalone "--") and accepts flag=value
func ExtraKVArgs(args []string) (extras map[string]string) {
	extras = make(map[string]string)
	for _, v := range args {
		split := strings.SplitN(v, "=", 2)
		switch {
		case len(split) == 2:
			extras[split[0]] = split[1]
		case len(split) == 1:
			extras[split[0]] = ""
		}
	}
	return
}

