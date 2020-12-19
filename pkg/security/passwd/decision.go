package passwd

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"time"
)

const (
	MessageAccountDisabled = "Account Disabled"
	MessageAccountLocked = "Account Locked"
	MessagePasswordLoginNotAllowed = "Password Login not Allowed"
)

/******************************
	abstracts
 ******************************/
type AuthenticationDecisionMaker interface {
	// Decide makes decision on whether the Authenticator should approve the auth request
	// returned error indecate the reason of rejection. returns nil when approved
	// 	 - The security.Authentication is nil when credentials has not been validated (pre check)
	// 	 - The security.Authentication is non-nil when credentials has been validated (post check).
	//     The non-nil is the proposed authentication to be returned by Authenticator
	//
	// If any of input paramters are mutable, AuthenticationDecisionMaker is allowed to change it
	Decide(context.Context, security.Candidate, security.Account, security.Authentication) error
}

type StageConditionFunc func(context.Context, security.Candidate, security.Account, security.Authentication) bool

type StageAwareDecisionMaker struct {
	delegate AuthenticationDecisionMaker
	condition StageConditionFunc
}

func (dm *StageAwareDecisionMaker) Decide(ctx context.Context, c security.Candidate, acct security.Account, auth security.Authentication) error {
	if dm.delegate == nil || dm.condition != nil && !dm.condition(ctx, c, acct, auth) {
		return nil
	}
	return dm.delegate.Decide(ctx, c, acct, auth)
}

func PreCheck(delegate AuthenticationDecisionMaker) AuthenticationDecisionMaker {
	return &StageAwareDecisionMaker{
		delegate: delegate,
		condition: isPreCheck,
	}
}

func PostCheck(delegate AuthenticationDecisionMaker) AuthenticationDecisionMaker {
	return &StageAwareDecisionMaker{
		delegate: delegate,
		condition: isPostCheck,
	}
}

func FinalCheck(delegate AuthenticationDecisionMaker) AuthenticationDecisionMaker {
	return &StageAwareDecisionMaker{
		delegate: delegate,
		condition: isFinalStage,
	}
}

/******************************
	helpers
 ******************************/
func performChecks(checkers []AuthenticationDecisionMaker, ctx context.Context, can security.Candidate, acct security.Account, auth security.Authentication) error {
	for _, checker := range checkers {
		if err := checker.Decide(ctx, can, acct, auth); err != nil {
			return err
		}
	}
	return nil
}

func isPreCheck(_ context.Context, _ security.Candidate, _ security.Account, auth security.Authentication) bool {
	return auth == nil
}

func isPostCheck(_ context.Context, _ security.Candidate, _ security.Account, auth security.Authentication) bool {
	return auth != nil
}

func isPreMFA(_ context.Context, can security.Candidate, _ security.Account, auth security.Authentication) bool {
	if auth != nil {
		return false
	}

	if _, isMFAVerify := can.(*MFAOtpVerification); isMFAVerify {
		return true
	}

	_, isMFARefresh := can.(*MFAOtpRefresh)
	return isMFARefresh
}

func isPostMFA(_ context.Context, can security.Candidate, _ security.Account, auth security.Authentication) bool {
	if auth == nil {
		return false
	}

	if _, isMFAVerify := can.(*MFAOtpVerification); isMFAVerify {
		return true
	}

	_, isMFARefresh := can.(*MFAOtpRefresh)
	return isMFARefresh
}

func isFinalStage(_ context.Context, can security.Candidate, _ security.Account, auth security.Authentication) bool {
	return auth != nil && auth.State() >= security.StateAuthenticated
}

/******************************
	Common Checks
 ******************************/
// AccountStatusChecker check account status and also auto unlock account if locking rules allows
type AccountStatusChecker struct {
	store security.AccountStore
}

func NewAccountStatusChecker(store security.AccountStore) *AccountStatusChecker {
	return &AccountStatusChecker{store: store}
}

func (adm *AccountStatusChecker) Decide(ctx context.Context, _ security.Candidate, acct security.Account, auth security.Authentication) error {
	switch {
	case acct.Disabled():
		return security.NewAccountStatusError(MessageAccountDisabled)
	case acct.Type() != security.AccountTypeDefault:
		return security.NewAccountStatusError(MessagePasswordLoginNotAllowed)
	case acct.Locked():
		return adm.decideAutoUnlock(ctx, acct)
	default:
		return nil
	}
}

func (adm *AccountStatusChecker) decideAutoUnlock(ctx context.Context, acct security.Account) (err error) {
	if !acct.Locked() {
		return nil
	}

	err = security.NewAccountStatusError(MessageAccountLocked)

	history, hok := acct.(security.AccountHistory)
	updater, uok := acct.(security.AccountUpdater)
	if !hok || !uok || history.LockoutTime().IsZero() {
		return
	}

	rules, err := adm.store.LoadLockingRules(ctx, acct)
	if err != nil || !rules.LockoutEnabled() || rules.LockoutDuration() <= 0 {
		return
	}

	if time.Now().After(history.LockoutTime().Add(rules.LockoutDuration()) ) {
		updater.Unlock()
	}

	if !acct.Locked() {
		return nil
	}

	return
}

// PasswordPolicyChecker takes account password policy into consideration
type PasswordPolicyChecker struct {
	store security.AccountStore
}

func NewPasswordPolicyChecker(store security.AccountStore) *PasswordPolicyChecker {
	return &PasswordPolicyChecker{store: store}
}

func (adm *PasswordPolicyChecker) Decide(_ context.Context, _ security.Candidate, acct security.Account, auth security.Authentication) error {
	return nil
}

