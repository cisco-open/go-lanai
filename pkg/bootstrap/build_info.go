package bootstrap

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"encoding/json"
	"path"
	"strings"
	"time"
)

var (
	// to be overridden by -ldflags

	BuildVersion = "Unknown"
	BuildTime    = time.Now().Format(utils.ISO8601Seconds)
	BuildHash    = "Unknown"
	BuildDeps    = "cto-github.cisco.com/NFV-BU/go-lanai@develop"
)

var (
	BuildInfo = BuildInfoMetadata{
		Version: BuildVersion,
		BuildTime: utils.ParseTime(utils.ISO8601Seconds, BuildTime),
		Hash: BuildHash,
		Modules: ModuleBuildInfoMap{},
	}

	BuildInfoMap map[string]interface{}
)

func init() {
	_ = (&BuildInfo.Modules).UnmarshalText([]byte(BuildDeps))
	BuildInfoMap = BuildInfo.ToMap()
}

type ModuleBuildInfo struct {
	Path    string `json:"path"`
	Version string `json:"version"`
}

type ModuleBuildInfoMap map[string]ModuleBuildInfo

func (m *ModuleBuildInfoMap) UnmarshalText(text []byte) error {
	mods := strings.Split(string(text), ",")
	modules := ModuleBuildInfoMap{}
	for _, v := range mods {
		tokens := strings.SplitN(strings.TrimSpace(v), "@", 2)
		if len(tokens) < 2 {
			continue
		}
		name := path.Base(tokens[0])
		modules[name] = ModuleBuildInfo{
			Path: tokens[0],
			Version: tokens[1],
		}
	}
	*m = modules
	return nil
}

type BuildInfoMetadata struct {
	Version   string    `json:"version"`
	BuildTime time.Time `json:"build-time"`
	Hash      string	`json:"hash"`
	Modules   ModuleBuildInfoMap `json:"modules,omitempty"`
}

func (m *BuildInfoMetadata) ToMap() map[string]interface{} {
	data, e := json.Marshal(m)
	if e != nil {
		return map[string]interface{}{}
	}

	ret := map[string]interface{}{}
	if e := json.Unmarshal(data, &ret); e != nil {
		return map[string]interface{}{}
	}
	return ret
}
