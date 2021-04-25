package bootstrap

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"encoding/json"
	"time"
)

var (
	BuildVersion = "Unknown"
	BuildTime    string
	BuildHash    string

	BuildInfo = BuildInfoMetadata{
		Version: BuildVersion,
		BuildTime: utils.ParseTime(utils.ISO8601Seconds, BuildTime),
		Hash: BuildHash,
		Modules: map[string]ModuleBuildInfo{},
	}

	BuildInfoMap = BuildInfo.ToMap()
)

type ModuleBuildInfo struct {
	Path    string `json:"path"`
	Version string `json:"version"`
}

type BuildInfoMetadata struct {
	Version   string    `json:"version"`
	BuildTime time.Time `json:"build-time"`
	Hash      string	`json:"hash"`
	Modules   map[string]ModuleBuildInfo `json:"modules,omitempty"`
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
