package tokenauth

/******************************
	Serialization
******************************/
func GobRegister() {

}

/************************
	security.Candidate
************************/

// BearerToken is the supported security.Candidate of resource server authenticator
type BearerToken struct {
	Token string
	DetailsMap map[string]interface{}
}

func (t *BearerToken) Principal() interface{} {
	return ""
}

func (t *BearerToken) Credentials() interface{} {
	return t.Token
}

func (t *BearerToken) Details() interface{} {
	return t.DetailsMap
}
