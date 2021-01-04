package csrf

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/mock_session"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCsrfMiddlewareShouldGenerateToken(t *testing.T) {
	csrfStore := newSessionBackedStore()
	manager := newManager(csrfStore, nil, nil)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSessionStore := mock_session.NewMockStore(ctrl)

	mockSessionStore.EXPECT().Options().Return(&session.Options{})
	s := session.NewSession(mockSessionStore, session.DefaultName)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set(web.ContextKeySession, s)
	c.Request = httptest.NewRequest("GET", "/form", nil)

	mockSessionStore.EXPECT().Save(gomock.Any()).Do(func(s *session.Session) {
		savedCsrfToken := s.Get(SessionKeyCsrfToken)
		if savedCsrfToken == nil {
			t.Errorf("Expect the csrf token to be saved in the session")
		}
		if savedCsrfToken.(*Token).ParameterName != "_csrf" {
			t.Errorf("Expected parameter name to be _csrf, but was %v", savedCsrfToken.(Token).ParameterName)
		}
		if savedCsrfToken.(*Token).HeaderName != "X-CSRF-TOKEN" {
			t.Errorf("Expected header name to be X-CSRF-TOKEN, but was %v", savedCsrfToken.(Token).HeaderName)
		}
		if savedCsrfToken.(*Token).Value == "" {
			t.Errorf("Expect csrf token value to not be empty")
		}

	})
	mw := manager.CsrfHandlerFunc()
	mw(c)

	csrfToken , ok := c.Get(web.ContextKeyCsrf)

	if !ok || csrfToken == ""{
		t.Errorf("expected to have session")
	}
}

func TestCsrfMiddlewareShouldCheckToken(t *testing.T) {
	csrfStore := newSessionBackedStore()
	manager := newManager(csrfStore, nil, nil)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSessionStore := mock_session.NewMockStore(ctrl)

	mockSessionStore.EXPECT().Options().Return(&session.Options{})
	s := session.NewSession(mockSessionStore, session.DefaultName)

	//Request with a invalid csrf token
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	existingCsrfToken := csrfStore.Generate(c, "_csrf", "X-CSRF-TOKEN")
	s.Set(SessionKeyCsrfToken, existingCsrfToken)

	c.Set(web.ContextKeySession, s)
	r := httptest.NewRequest("POST", "/process", strings.NewReader("_csrf=" + uuid.New().String()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.Request = r
	mw := manager.CsrfHandlerFunc()
	mw(c)

	if len(c.Errors) != 1 {
		t.Errorf("there should be one error")
	}

	if !errors.Is(c.Errors.Last().Err, security.NewInvalidCsrfTokenError("")) {
		t.Errorf("expect invalid csrf token error, but was %v", c.Errors.Last())
	}

	//Request without csrf token
	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Set(web.ContextKeySession, s)
	r = httptest.NewRequest("POST", "/process", strings.NewReader("a=b"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.Request = r

	mw(c)

	if len(c.Errors) != 1 {
		t.Errorf("there should be one error")
	}

	if !errors.Is(c.Errors.Last().Err, security.NewMissingCsrfTokenError("")) {
		t.Errorf("expect missing csrf token error, but was %v", c.Errors.Last())
	}

	//Request with expected csrf token
	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Set(web.ContextKeySession, s)
	r = httptest.NewRequest("POST", "/process", strings.NewReader("_csrf=" + existingCsrfToken.Value))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.Request = r

	mw(c)

	if len(c.Errors) != 0 {
		t.Errorf("there should be no error")
	}

	//Request with expected csrf token in header instead of form parameter
	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Set(web.ContextKeySession, s)
	r = httptest.NewRequest("POST", "/process", nil)
	r.Header.Set("X-CSRF-TOKEN", existingCsrfToken.Value)
	c.Request = r

	mw(c)

	if len(c.Errors) != 0 {
		t.Errorf("there should be no error")
	}

	//Request without a csrf token associated with it
	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	s.Delete(SessionKeyCsrfToken) //remove the csrf token from the session
	c.Set(web.ContextKeySession, s)
	r = httptest.NewRequest("POST", "/process", nil)
	r.Header.Set("X-CSRF-TOKEN", uuid.New().String())
	c.Request = r

	mockSessionStore.EXPECT().Save(gomock.Any()) //since this request's session doesn't have a csrf token, one will be generated

	mw(c)

	if len(c.Errors) != 1 {
		t.Errorf("there should be one error")
	}

	if !errors.Is(c.Errors.Last().Err, security.NewInvalidCsrfTokenError("")) {
		t.Errorf("expect invalid csrf token error, but was %v", c.Errors.Last())
	}
}