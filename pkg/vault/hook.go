package vault

import "context"

type Hook interface {
	BeforeOperation(ctx context.Context, cmd string) context.Context
	AfterOperation(ctx context.Context, err error)
}
