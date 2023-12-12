package aws

import (
    "cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
    "go.uber.org/fx"
)

var Module = &bootstrap.Module{
    Name:       "AWS",
    Precedence: bootstrap.AwsPrecedence,
    Options: []fx.Option{
        fx.Provide(BindAwsProperties, NewConfigLoader),
    },
}
