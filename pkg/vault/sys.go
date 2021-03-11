package vault

import (
	"context"
	"github.com/hashicorp/vault/api"
)

type Sys struct {
	*api.Sys
	ctx context.Context
	hooks []Hook
}