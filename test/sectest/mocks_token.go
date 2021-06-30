package sectest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
)

const (
	tokenDelimiter = "~"
)

/*************************
	Token
 *************************/

type MockedTokenInfo struct {
	UName string
	UID   string
	TID   string
	TName string
	OrigU string
	Exp   int64
	Iss   int64
}

// MockedToken implements oauth2.AccessToken
type MockedToken struct {
	MockedTokenInfo
	ExpTime time.Time `json:"-"`
	IssTime time.Time `json:"-"`
}

func (mt MockedToken) MarshalText() (text []byte, err error) {
	mt.Exp = mt.ExpTime.UnixNano()
	mt.Iss = mt.IssTime.UnixNano()
	text, err = json.Marshal(mt.MockedTokenInfo)
	if err != nil {
		return
	}
	return []byte(base64.StdEncoding.EncodeToString(text)), nil
}

func (mt *MockedToken) UnmarshalText(text []byte) error {
	data, e := base64.StdEncoding.DecodeString(string(text))
	if e != nil {
		return e
	}
	if e := json.Unmarshal(data, &mt.MockedTokenInfo); e != nil {
		return e
	}
	mt.ExpTime = time.Unix(0, mt.Exp)
	mt.IssTime = time.Unix(0, mt.Iss)
	return nil
}

func (mt MockedToken) String() string {
	vals := []string{mt.UName, mt.UID, mt.TID, mt.TName, mt.OrigU, mt.ExpTime.Format(utils.ISO8601Milliseconds)}
	return strings.Join(vals, tokenDelimiter)
}

func (mt *MockedToken) Value() string {
	text, e := mt.MarshalText()
	if e != nil {
		return ""
	}
	return string(text)
}

func (mt *MockedToken) ExpiryTime() time.Time {
	return mt.ExpTime
}

func (mt *MockedToken) Expired() bool {
	return !mt.ExpTime.IsZero() && !time.Now().Before(mt.ExpTime)
}

func (mt *MockedToken) Details() map[string]interface{} {
	return map[string]interface{}{}
}

func (mt *MockedToken) Type() oauth2.TokenType {
	return oauth2.TokenTypeBearer
}

func (mt *MockedToken) IssueTime() time.Time {
	return mt.IssTime
}

func (mt *MockedToken) Scopes() utils.StringSet {
	return utils.NewStringSet()
}

func (mt *MockedToken) RefreshToken() oauth2.RefreshToken {
	return nil
}
