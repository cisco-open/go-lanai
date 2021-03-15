package data

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"github.com/pkg/errors"
	"time"
)

const (
	ManagementPropertiesPrefix = "data"
)

type DataProperties struct {
	Logging     LoggingProperties `json:"logging"`
}

type LoggingProperties struct {
	Level         log.LoggingLevel `json:"level"`
	SlowThreshold utils.Duration   `json:"slow-threshold"`
}

//NewDataProperties create a DataProperties with default values
func NewDataProperties() *DataProperties {
	return &DataProperties{
		Logging:     LoggingProperties{
			Level: log.LevelWarn,
			SlowThreshold: utils.Duration(15 * time.Second),
		},
	}
}

//BindDataProperties create and bind SessionProperties, with a optional prefix
func BindDataProperties(ctx *bootstrap.ApplicationContext) DataProperties {
	props := NewDataProperties()
	if err := ctx.Config().Bind(props, ManagementPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind DataProperties"))
	}
	return *props
}


