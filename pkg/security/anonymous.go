package security

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

func (_ *AnonymousAuthentication) Permissions() []string {
	return []string{}
}

func (_ *AnonymousAuthentication) Authenticated() bool {
	return false
}

func (aa *AnonymousAuthentication) Details() interface{} {
	return aa.Details()
}

type AnonymousAuthenticator struct{}

func (a *AnonymousAuthenticator) Authenticate(candidate Candidate) (auth Authentication, err error) {
	if ac, ok := candidate.(AnonymousCandidate); ok {
		return &AnonymousAuthentication{candidate: ac}, nil
	}
	return nil, nil
}
