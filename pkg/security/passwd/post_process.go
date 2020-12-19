package passwd

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
)

/******************************
	abstracts
 ******************************/
// AuthenticationDecision is a values carrier for PostAuthenticationProcessor
type AuthenticationDecision struct {
	Candidate security.Candidate
	Auth      security.Authentication
	Error     error
}

// PostAuthenticationProcessor is invoked at the end of authentication process regardless of authentication decisions (granted or rejected)
// If PostAuthenticationProcessor implement order.Ordered interface, its order is respected using order.OrderedFirstCompareReverse.
// This means highest priority is executed last
type PostAuthenticationProcessor interface {
	// Process is invoked at the end of authentication process by the Authenticator to perform post-auth action.
	// The method is invoked regardless if the authentication is granted:
	// 	- If the authentication is granted, the AuthenticationDecision.Auth is non-nil and AuthenticationDecision.Error is nil
	//	- If the authentication is rejected, the AuthenticationDecision.Error is non-nil and AuthenticationDecision.Auth should be ignored
	//
	// If the context.Context and security.Account paramters are mutable, PostAuthenticationProcessor is allowed to change it
	// Note: PostAuthenticationProcessor cannot overwrite authentication decision.
	Process(context.Context, security.Account, AuthenticationDecision)
}

/******************************
	Helpers
 ******************************/
func applyPostAuthenticationProcessors(processors []PostAuthenticationProcessor,
	ctx context.Context, acct security.Account,
	can security.Candidate, auth security.Authentication, err error) {

	decision := AuthenticationDecision {
		Candidate: can,
		Auth: auth,
		Error: err,
	}
	for _,processor := range processors {
		processor.Process(ctx, acct, decision)
	}
}

/******************************
	Common Implementation
 ******************************/
// PersistAccountPostProcessor saves Account. It's implement order.Ordered with highest priority
// Note: post-processors executed in reversed order
type PersistAccountPostProcessor struct {
	store security.AccountStore
}

func NewPersistAccountPostProcessor(store security.AccountStore) *PersistAccountPostProcessor {
	return &PersistAccountPostProcessor{store: store}
}

func (p *PersistAccountPostProcessor) Process(ctx context.Context, acct security.Account, d AuthenticationDecision) {
	if acct == nil {
		return
	}

	// regardless decision, account need to be persisted in case of any status changes.
	// Note: we ignore save error since it's too late to do anything
	_ = p.store.Save(ctx, acct)
}
