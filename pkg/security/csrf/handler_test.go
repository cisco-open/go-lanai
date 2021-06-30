package csrf

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session/common"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/mocks/authmock"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/mocks/sessionmock"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"net/http/httptest"
	"testing"
)

func TestChangeCsrfHanlderShouldChangeCSRFTokenWhenAuthenticated(t *testing.T) {
	csrfStore := newSessionBackedStore()
	handler := &ChangeCsrfHanlder{
		csrfStore,
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSessionStore := sessionmock.NewMockStore(ctrl)

	mockSessionStore.EXPECT().Options().Return(&session.Options{})
	s := session.NewSession(mockSessionStore, common.DefaultName)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set(web.ContextKeySession, s)
	token := &Token{
		uuid.New().String(),
		security.CsrfParamName,
		security.CsrfHeaderName,
	}
	s.Set(SessionKeyCsrfToken, token)
	//The request itself is not important
	c.Request = httptest.NewRequest("GET", "/something", nil)

	mockFrom := authmock.NewMockAuthentication(ctrl)
	mockFrom.EXPECT().State().Return(security.StateAnonymous)

	mockTo := authmock.NewMockAuthentication(ctrl)
	mockTo.EXPECT().State().Return(security.StateAuthenticated)

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
		if savedCsrfToken.(*Token).Value == token.Value {
			t.Errorf("Expect csrf token value should change")
		}
	})

	handler.HandleAuthenticationSuccess(c, c.Request, c.Writer, mockFrom, mockTo)
}

func TestChangeCsrfHanlderShouldNotChangeCSRFTokenIfNotAuthenticated(t *testing.T) {
	csrfStore := newSessionBackedStore()
	handler := &ChangeCsrfHanlder{
		csrfStore,
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSessionStore := sessionmock.NewMockStore(ctrl)

	mockSessionStore.EXPECT().Options().Return(&session.Options{})
	s := session.NewSession(mockSessionStore, common.DefaultName)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set(web.ContextKeySession, s)
	token := &Token{
		uuid.New().String(),
		security.CsrfParamName,
		security.CsrfHeaderName,
	}
	s.Set(SessionKeyCsrfToken, token)
	//The request itself is not important
	c.Request = httptest.NewRequest("GET", "/something", nil)

	mockFrom := authmock.NewMockAuthentication(ctrl)
	mockFrom.EXPECT().State().Return(security.StateAuthenticated)

	mockTo := authmock.NewMockAuthentication(ctrl)
	mockTo.EXPECT().State().Return(security.StateAnonymous)

	handler.HandleAuthenticationSuccess(c, c.Request, c.Writer, mockFrom, mockTo)

	actualToken := s.Get(SessionKeyCsrfToken).(*Token)
	if actualToken.Value != token.Value {
		t.Errorf("csrf token should be unchanged")
	}
}