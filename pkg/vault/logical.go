package vault

import (
	"context"
	"github.com/hashicorp/vault/api"
)

type Logical struct {
	*api.Logical
	ctx context.Context
	hooks []Hook
}

func (l *Logical) Read(path string) (*api.Secret, error) {
	for _, h := range l.hooks {
		h.BeforeOperation(l.ctx, "Read")
	}

	secrets, err := l.Logical.Read(path)

	for _, h := range l.hooks {
		h.AfterOperation(l.ctx, err)
	}
	return secrets, err
}