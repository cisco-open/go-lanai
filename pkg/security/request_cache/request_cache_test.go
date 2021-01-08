package request_cache

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/mock_redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/mock_security"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/mock_session"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestSaveAndGetCachedRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRedis := mock_redis.NewMockUniversalClient(ctrl)

	sessionStore := session.NewRedisStore(mockRedis)
	s, _ := sessionStore.New(session.DefaultName)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Set(web.ContextKeySession, s)
	c.Request = httptest.NewRequest("POST", "/something", strings.NewReader("a=b&c=d"))
	c.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	SaveRequest(c)
	cached := GetCachedRequest(c)

	if cached.Method != "POST" {
		t.Errorf("expected POST, but got %s", cached.Method)
	}

	if cached.PostForm.Get("a") != "b" && cached.PostForm.Get("c") != "d" {
		t.Errorf("expected post form to have a=b and c=d")
	}
}


func TestCachedRequestPreProcessor_Process(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSessionStore := mock_session.NewMockStore(ctrl)

	processor := newCachedRequestPreProcessor(mockSessionStore)

	mockSessionStore.EXPECT().Options().Return(&session.Options{})

	s := session.CreateSession(mockSessionStore, session.DefaultName)

	cached := &CachedRequest{
		Host: "example.com",
		Method: "POST",
		URL: &url.URL{Path: "/something"},
		PostForm: url.Values{"a":[]string{"b"},"c":[]string{"d"}},
	}

	s.Set(SessionKeyCachedRequest, cached)
	mockSessionStore.EXPECT().Get(s.GetID(), s.Name()).Return(s, nil)
	mockSessionStore.EXPECT().Save(s).Do(func(session *session.Session) {
		if session.Get(SessionKeyCachedRequest) != nil {
			t.Errorf("cached request should be removed from the session")
		}
	})

	//GET request to the same path
	req := httptest.NewRequest("GET", "/something", nil)
	req.Header.Set("Cookie", session.DefaultName+"="+s.GetID())

	_ = processor.Process(req)

	if req.Method != "POST" {
		t.Errorf("expect the method to be changed to match the cached request")
	}

	if req.PostForm.Get("a") != "b" && req.PostForm.Get("c") != "d" {
		t.Errorf("expected post form to have cached value")
	}
}

func TestSavedRequestAuthenticationSuccessHandler_HandleAuthenticationSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRedis := mock_redis.NewMockUniversalClient(ctrl)

	sessionStore := session.NewRedisStore(mockRedis)
	s, _ := sessionStore.New(session.DefaultName)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Set(web.ContextKeySession, s)
	c.Request = httptest.NewRequest("POST", "/something", strings.NewReader("a=b&c=d"))
	c.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	SaveRequest(c)

	mockFrom := mock_security.NewMockAuthentication(ctrl)
	mockFrom.EXPECT().State().Return(security.StateAnonymous)

	mockTo := mock_security.NewMockAuthentication(ctrl)
	mockTo.EXPECT().State().Return(security.StateAuthenticated)

	handler := NewSavedRequestAuthenticationSuccessHandler(nil)

	handler.HandleAuthenticationSuccess(c, c.Request, c.Writer, mockFrom, mockTo)

	if recorder.Result().StatusCode != 302 {
		t.Errorf("expected 302 but got %v ", recorder.Result().StatusCode )
	}

	l, _ := recorder.Result().Location()

	if l.String() != "/something" {
		t.Errorf("expected redirect location, got %v", l)
	}
}

func TestSaveRequestEntryPoint_Commence(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRedis := mock_redis.NewMockUniversalClient(ctrl)

	sessionStore := session.NewRedisStore(mockRedis)
	s, _ := sessionStore.New(session.DefaultName)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Set(web.ContextKeySession, s)
	c.Set(web.ContextKeyContextPath, "")

	entryPoint := NewSaveRequestEntryPoint(&noOpEntryPoint{})

	c.Request = httptest.NewRequest("GET", "/something/favicon.jpg", nil)
	entryPoint.Commence(c, c.Request, c.Writer, security.NewAccessDeniedError("access denied"))
	if GetCachedRequest(c) != nil {
		t.Errorf("request for favicon should not be cached")
	}

	s, _ = sessionStore.New(session.DefaultName)
	recorder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recorder)
	c.Set(web.ContextKeySession, s)
	c.Set(web.ContextKeyContextPath, "")
	c.Request = httptest.NewRequest("GET", "/something", nil)
	c.Request.Header.Set("X-Requested-With", "XMLHttpRequest")
	entryPoint.Commence(c, c.Request, c.Writer, security.NewAccessDeniedError("access denied"))
	if GetCachedRequest(c) != nil {
		t.Errorf("request with  XMLHttpRequest should not be cached")
	}

	s, _ = sessionStore.New(session.DefaultName)
	recorder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recorder)
	c.Set(web.ContextKeySession, s)
	c.Set(web.ContextKeyContextPath, "")
	c.Request = httptest.NewRequest("GET", "/something", nil)
	c.Request.Header.Set("Trailer", "anything")
	entryPoint.Commence(c, c.Request, c.Writer, security.NewAccessDeniedError("access denied"))
	if GetCachedRequest(c) != nil {
		t.Errorf("request with  XMLHttpRequest should not be cached")
	}

	s, _ = sessionStore.New(session.DefaultName)
	recorder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recorder)
	c.Set(web.ContextKeySession, s)
	c.Set(web.ContextKeyContextPath, "")
	c.Request = httptest.NewRequest("GET", "/something", nil)
	c.Request.Header.Set("Content-Type", "multipart/form-data something")
	entryPoint.Commence(c, c.Request, c.Writer, security.NewAccessDeniedError("access denied"))
	if GetCachedRequest(c) != nil {
		t.Errorf("request with multipart/form-data should not be cached")
	}

	s, _ = sessionStore.New(session.DefaultName)
	recorder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recorder)
	c.Set(web.ContextKeySession, s)
	c.Set(web.ContextKeyContextPath, "")
	c.Request = httptest.NewRequest("POST", "/something", strings.NewReader("a=b&c=d"))
	c.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.Request.Header.Set(security.CsrfHeaderName, "something")
	entryPoint.Commence(c, c.Request, c.Writer, security.NewAccessDeniedError("access denied"))
	if GetCachedRequest(c) != nil {
		t.Errorf("request with csrf header should not be cached")
	}

	s, _ = sessionStore.New(session.DefaultName)
	recorder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recorder)
	c.Set(web.ContextKeySession, s)
	c.Set(web.ContextKeyContextPath, "")
	c.Request = httptest.NewRequest("POST", "/something", strings.NewReader(security.CsrfParamName + "=something"))
	c.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	entryPoint.Commence(c, c.Request, c.Writer, security.NewAccessDeniedError("access denied"))
	if GetCachedRequest(c) != nil {
		t.Errorf("request with csrf param should not be cached")
	}

	s, _ = sessionStore.New(session.DefaultName)
	recorder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recorder)
	c.Set(web.ContextKeySession, s)
	c.Set(web.ContextKeyContextPath, "")
	c.Request = httptest.NewRequest("POST", "/something", strings.NewReader("a=b&c=d"))
	c.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	entryPoint.Commence(c, c.Request, c.Writer, security.NewAccessDeniedError("access denied"))

	if GetCachedRequest(c) == nil {
		t.Errorf("expect request to be cached")
	}
}

type noOpEntryPoint struct {}
func (e *noOpEntryPoint) Commence(context.Context, *http.Request, http.ResponseWriter, error) {}

