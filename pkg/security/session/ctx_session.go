package session

import "context"

type SettingService interface {
	GetMaximumSessions(ctx context.Context) int
}