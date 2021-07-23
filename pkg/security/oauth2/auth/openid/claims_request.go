package openid

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/claims"
	"encoding/json"
)

type claimsRequest struct {
	UserInfo requestedClaims `json:"userinfo"`
	IdToken  requestedClaims `json:"id_token"`
}

// requestedClaims implements claims.RequestedClaims
type requestedClaims map[string]requestedClaim

func (r requestedClaims) Get(claim string) (c claims.RequestedClaim, ok bool) {
	c, ok = r[claim]
	return
}

type rcDetails struct {
	Essential   bool     `json:"essential"`
	Values      []string `json:"values,omitempty"`
	SingleValue *string  `json:"value,omitempty"`
}

// requestedClaim implements claims.RequestedClaim and json.Unmarshaler
type requestedClaim struct {
	rcDetails
}

func (r requestedClaim) Essential() bool {
	return r.rcDetails.Essential
}

func (r requestedClaim) Values() []string {
	return r.rcDetails.Values
}

func (r requestedClaim) IsDefault() bool {
	return len(r.rcDetails.Values) == 0
}

func (r *requestedClaim) UnmarshalJSON(data []byte) error {
	r.rcDetails.Values = []string{}
	if e := json.Unmarshal(data, &r.rcDetails); e != nil {
		return e
	}

	if r.rcDetails.SingleValue != nil {
		r.rcDetails.Values = []string{*r.rcDetails.SingleValue}
		r.rcDetails.SingleValue = nil
	}
	return nil
}
