package csrf

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"encoding/gob"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const csrfTokenAttrName = "CsrfToken"

/**
 The header name and parameter name are part of the token in case some components down the line needs them.
 For example, if the token is used as a hidden variable in a form, the parameter name would be needed.
 */
type Token struct {
	Value string

	// the HTTP parameter that the CSRF token can be placed on request
	ParameterName string

	// the HTTP header that the CSRF can be placed on requests instead of the parameter.
	HeaderName string
}

/**
	The store is responsible for reading the CSRF token associated to the request.
	How the CSRF token is associated to the request is the implementation's discretion.

	The store is responsible for writing to the response header if necessary
	for example, if the store implementation is based on cookies, then the save method
	would write (save) the token as a cookie header.
 */
type TokenStore interface {
	Generate(c *gin.Context, parameterName string, headerName string) *Token

	SaveToken(c *gin.Context, token *Token) error

	LoadToken(c *gin.Context) (*Token, error)
}

type SessionBackedStore struct {
}

func newSessionBackedStore() *SessionBackedStore{
	gob.Register((*Token)(nil))
	return &SessionBackedStore{}
}

func (store *SessionBackedStore) Generate(_ *gin.Context, parameterName string, headerName string) *Token {
	t := &Token{
		Value: uuid.New().String(),
		ParameterName: parameterName,
		HeaderName: headerName,
	}
	return t
}

func (store *SessionBackedStore) SaveToken(c *gin.Context, token *Token) error {
	s := session.Get(c)

	if s == nil {
		return errors.New("can't save csrf token to session, because the request has no session")
	}

	s.Set(csrfTokenAttrName, token)
	return s.Save()
}

func (store *SessionBackedStore) LoadToken(c *gin.Context) (*Token, error) {
	s := session.Get(c)

	if s == nil {
		return nil, errors.New("can't load csrf token from session, because the request has no session")
	}

	attr := s.Get(csrfTokenAttrName)

	if attr == nil {
		return nil, nil
	}

	if token, ok := attr.(*Token); !ok {
		return nil, errors.New("csrf token in session can't be asserted to be the CSRF token type")
	} else {
		return token, nil
	}
}