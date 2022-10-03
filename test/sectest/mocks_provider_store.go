package sectest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"fmt"
)

const (
	MockedProviderID = "test-provider"
	MockedProviderName = "test-provider"
)

type MockedProviderStore struct {}

func (s MockedProviderStore) LoadProviderById(_ context.Context, id string) (*security.Provider, error) {
	if id != MockedProviderID {
		return nil, fmt.Errorf("cannot find provider with id [%s]", id)
	}
	return &security.Provider{
		Id:               id,
		Name:             MockedProviderName,
		DisplayName:      MockedProviderName,
		Description:      MockedProviderName,
		LocaleCode:       "en_US",
		NotificationType: "EMAIL",
		Email:            "admin@cisco.com",
	}, nil
}
