package infratest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
)

var logger = log.New("T.Infra")

const (
	_ = order.Highest + iota * 100
	orderEmbeddedRedis
)

