package passwd

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"fmt"
	"time"
)

/******************************
	abstracts
 ******************************/

// AuthenticationDecisionMaker is invoked at various stages of authentication decision making process.
// If AuthenticationDecisionMaker implement order.Ordered interface, its order is respected using order.OrderedFirstCompare.
// This means highest priority is executed first and non-ordered decision makers run at last.
//
// Note: each AuthenticationDecisionMaker will get invoked multiple times during the authentication process.
// 		 So implementations should check stage before making desisions. Or use ConditionalDecisionMaker
type AuthenticationDecisionMaker interface {
	// Decide makes decision on whether the Authenticator should approve the auth request.
	// the returned error indicate the reason of rejection. returns nil when approved
	// 	 - The security.Authentication is nil when credentials has not been validated (pre check)
	// 	 - The security.Authentication is non-nil when credentials has been validated (post check).
	//     The non-nil value is the proposed authentication to be returned by Authenticator
	//
	// If any of input paramters are mutable, AuthenticationDecisionMaker is allowed to change it
	Decide(context.Context, security.Candidate, security.Account, security.Authentication) error
}

/******************************
	Common Implementation
 ******************************/

type DecisionMakerConditionFunc func(context.Context, security.Candidate, security.Account, security.Authentication) bool

// ConditionalDecisionMaker implements AuthenticationDecisionMaker with ability to skip based on condiitons
type ConditionalDecisionMaker struct {
	delegate  AuthenticationDecisionMaker
	condition DecisionMakerConditionFunc
}

func (dm *ConditionalDecisionMaker) Decide(ctx context.Context, c security.Candidate, acct security.Account, auth security.Authentication) error {
	if dm.delegate == nil || dm.condition != nil && !dm.condition(ctx, c, acct, auth) {
		return nil
	}
	return dm.delegate.Decide(ctx, c, acct, auth)
}

func PreCredentialsCheck(delegate AuthenticationDecisionMaker) AuthenticationDecisionMaker {
	return &ConditionalDecisionMaker{
		delegate:  delegate,
		condition: isPreCredentialsCheck,
	}
}

func PostCredentialsCheck(delegate AuthenticationDecisionMaker) AuthenticationDecisionMaker {
	return &ConditionalDecisionMaker{
		delegate:  delegate,
		condition: isPostCredentialsCheck,
	}
}

func FinalCheck(delegate AuthenticationDecisionMaker) AuthenticationDecisionMaker {
	return &ConditionalDecisionMaker{
		delegate: delegate,
		condition: isFinalStage,
	}
}

/******************************
	helpers
 ******************************/
func makeDecision(checkers []AuthenticationDecisionMaker, ctx context.Context, can security.Candidate, acct security.Account, auth security.Authentication) error {
	for _, checker := range checkers {
		if err := checker.Decide(ctx, can, acct, auth); err != nil {
			return err
		}
	}
	return nil
}

func isPreCredentialsCheck(_ context.Context, _ security.Candidate, _ security.Account, auth security.Authentication) bool {
	return auth == nil
}

func isPostCredentialsCheck(_ context.Context, _ security.Candidate, _ security.Account, auth security.Authentication) bool {
	return auth != nil
}

func isPreMFAVerify(_ context.Context, can security.Candidate, _ security.Account, auth security.Authentication) bool {
	if auth != nil {
		return false
	}

	if _, isMFAVerify := can.(*MFAOtpVerification); isMFAVerify {
		return true
	}

	_, isMFARefresh := can.(*MFAOtpRefresh)
	return isMFARefresh
}

func isPostMFAVerify(_ context.Context, can security.Candidate, _ security.Account, auth security.Authentication) bool {
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

func (adm *AccountStatusChecker) Decide(ctx context.Context, _ security.Candidate, acct security.Account, _ security.Authentication) error {
	if acct == nil {
		return nil
	}

	switch {
	case acct.Disabled():
		return security.NewAccountStatusError(MessageAccountDisabled)
	case acct.Type() == security.AccountTypeFederated:
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

	rules, e := adm.store.LoadLockingRules(ctx, acct)
	if e != nil || rules == nil || !rules.LockoutEnabled() || rules.LockoutDuration() <= 0 {
		return
	}

	if time.Now().After(history.LockoutTime().Add(rules.LockoutDuration()) ) {
		updater.UnlockAccount()
		logger.WithContext(ctx).Infof("Account[%s] Unlocked", acct.Username())
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

func (c *PasswordPolicyChecker) Decide(ctx context.Context, _ security.Candidate, acct security.Account, auth security.Authentication) error {
	history, hok := acct.(security.AccountHistory)
	_, uok := acct.(security.AccountUpdater)
	if !hok || !uok {
		return nil
	}

	policy, e := c.store.LoadPwdAgingRules(ctx, acct)
	if e != nil || policy == nil || !policy.PwdAgingRuleEnforced() || policy.PwdMaxAge() <= 0 {
		return nil
	}

	switch {
	case history.PwdChangedTime().Add(policy.PwdMaxAge()).Before(time.Now()):
		return c.decideExpiredPassword(ctx, acct, policy, auth)
	default:
		return c.decideNonExpiredPassword(ctx, acct, policy, auth)
	}
}

func (c *PasswordPolicyChecker) decideNonExpiredPassword(
	_ context.Context, acct security.Account, policy security.AccountPwdAgingRule, auth security.Authentication) (err error) {

	// reset graceful auth
	acct.(security.AccountUpdater).ResetGracefulAuthCount()

	// check if expiring soon
	toExpire := policy.PwdMaxAge() - time.Now().Sub(acct.(security.AccountHistory).PwdChangedTime())
	if toExpire >= 0 && toExpire < policy.PwdExpiryWarningPeriod() {
		c.addWarning(auth, fmt.Sprintf("Password is expring in %s", toExpire.String()))
	}
	return nil
}

func (c *PasswordPolicyChecker) decideExpiredPassword(
	_ context.Context, acct security.Account, policy security.AccountPwdAgingRule, auth security.Authentication) error {

	switch remaining := policy.GracefulAuthLimit() - acct.(security.AccountHistory).GracefulAuthCount(); {
	case remaining <= 0:
		// No more graceful auth
		return security.NewCredentialsExpiredError(MessagePasswordExpired)
	case remaining == 1:
		// Last chance
		c.addWarning(auth, "Last Graceful Login")
	default:
		// more chance available
		c.addWarning(auth, fmt.Sprintf("%d Graceful Login Left", remaining))
	}

	acct.(security.AccountUpdater).IncrementGracefulAuthCount()
	return nil
}

func (c *PasswordPolicyChecker) addWarning(auth security.Authentication, warning interface{}) {
	details, ok := auth.Details().(map[string]interface{})
	if !ok || details == nil {
		return
	}

	var existing []interface{}
	switch w := details[security.DetailsKeyAuthWarning]; w.(type) {
	case nil:
		existing = []interface{}{}
	case []interface{}:
		existing = w.([]interface{})
	default:
		existing = []interface{}{w}
	}

	if warnings, ok := warning.([]interface{}); ok {
		details[security.DetailsKeyAuthWarning] = append(existing, warnings...)
	} else {
		details[security.DetailsKeyAuthWarning] = append(existing, warning)
	}
}

