package security

import "context"

type AnonymousCandidate map[string]interface{}

// security.Candidate
func (ac AnonymousCandidate) Principal() interface{} {
	return "anonymous"
}

// security.Candidate
func (_ AnonymousCandidate) Credentials() interface{} {
	return ""
}

// security.Candidate
func (ac AnonymousCandidate) Details() interface{} {
	return ac
}

type AnonymousAuthentication struct {
	candidate AnonymousCandidate
}

func (aa *AnonymousAuthentication) Principal() interface{} {
	return aa.candidate.Principal()
}

func (_ *AnonymousAuthentication) Permissions() map[string]interface{} {
	return map[string]interface{}{}
}

func (_ *AnonymousAuthentication) State() AuthenticationState {
	return StateAnonymous
}

func (aa *AnonymousAuthentication) Details() interface{} {
	return aa.Details()
}

type AnonymousAuthenticator struct{}

func (a *AnonymousAuthenticator) Authenticate(_ context.Context, candidate Candidate) (auth Authentication, err error) {
	if ac, ok := candidate.(AnonymousCandidate); ok {
		return &AnonymousAuthentication{candidate: ac}, nil
	}
	return nil, nil
}
