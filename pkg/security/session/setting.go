package session

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
)

type DefaultSettingService struct {
	sessionProperty security.SessionProperties
}

func NewDefaultSettingService(p security.SessionProperties) SettingService {
	return &DefaultSettingService{
		sessionProperty: p,
	}
}

func (d *DefaultSettingService) GetMaximumSessions(ctx context.Context) int {
	return d.sessionProperty.MaxConcurrentSession
}

