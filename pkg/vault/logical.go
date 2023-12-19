package vault

import (
	"context"
	"fmt"
	"github.com/hashicorp/vault/api"
	"io"
	"net/http"
	"strings"
)

type Logical struct {
	*api.Logical
	ctx context.Context
	client *Client
}

// WithContext make a copy of current Logical with a new context
func (l *Logical) WithContext(ctx context.Context) *Logical {
	if ctx == nil {
		panic("nil context is not allowed")
	}
	return &Logical{
		Logical: l.Logical,
		ctx:     ctx,
		client:  l.client,
	}
}

// Read override api.Logical with proper hooks
func (l *Logical) Read(path string) (ret *api.Secret, err error) {
	l.beforeOp("Read", path)
	defer func() { l.afterOp(err) }()

	ret, err = l.Logical.Read(path)
	return
}

// ReadWithData override api.Logical with proper hooks
// Note: data is sent as HTTP parameters
func (l *Logical) ReadWithData(path string, data map[string][]string) (ret *api.Secret, err error) {
	l.beforeOp("Read", path)
	defer func() { l.afterOp(err) }()

	ret, err = l.Logical.ReadWithData(path, data)
	return
}

// Write override api.Logical with proper hooks. This method accept data as an interface instead of map
// Note: Write sends PUT request
func (l *Logical) Write(path string, data interface{}) (ret *api.Secret, err error) {
	l.beforeOp("Write", path)
	defer func() { l.afterOp(err) }()

	ret, err = l.writeWithMethod(http.MethodPut, path, data) //nolint:contextcheck
	return
}

// Post is extension of api.Logical. Similar to Write, but use POST request
func (l *Logical) Post(path string, data interface{}) (ret *api.Secret, err error) {
	l.beforeOp("Post", path)
	defer func() { l.afterOp(err) }()

	ret, err = l.writeWithMethod(http.MethodPost, path, data) //nolint:contextcheck
	return
}

// WriteWithMethod is extension of api.Logical to send POST and PUT request
func (l *Logical) WriteWithMethod(method, path string, data interface{}) (ret *api.Secret, err error) {
	l.beforeOp(method, path)
	defer func() { l.afterOp(err) }()
	return l.writeWithMethod(strings.ToUpper(method), path, data) //nolint:contextcheck
}

func (l *Logical) beforeOp(name, path string) {
	cmd := fmt.Sprintf("%s %s", name, path)
	for _, h := range l.client.hooks {
		h.BeforeOperation(l.ctx, cmd)
	}
}

func (l *Logical) afterOp(err error) {
	for _, h := range l.client.hooks {
		h.AfterOperation(l.ctx, err)
	}
}

//nolint:contextcheck // context is bond with struct
func (l *Logical) writeWithMethod(method, path string, data interface{}) (*api.Secret, error) {
	switch method {
	case http.MethodPost, http.MethodPut:
	default:
		return nil, fmt.Errorf("invalid HTTP method, only POST and PUT are accepted")
	}

	r := l.client.NewRequest(method, "/v1/"+path)
	if e := r.SetJSONBody(data); e != nil {
		return nil, e
	}

	return l.write(r)
}

//nolint:contextcheck // context is bond with struct
func (l *Logical) write(request *api.Request) (*api.Secret, error) {
	ctx, cancelFunc := context.WithCancel(l.ctx)
	defer cancelFunc()
	//nolint:staticcheck // Deprecated API. TODO should fix
	resp, err := l.client.RawRequestWithContext(ctx, request)
	if resp != nil {
		defer resp.Body.Close()
	}
	if resp != nil && resp.StatusCode == 404 {
		secret, parseErr := api.ParseSecret(resp.Body)
		switch parseErr {
		case nil:
		case io.EOF:
			return nil, nil
		default:
			return nil, err
		}
		if secret != nil && (len(secret.Warnings) > 0 || len(secret.Data) > 0) {
			return secret, err
		}
	}
	if err != nil {
		return nil, err
	}

	return api.ParseSecret(resp.Body)
}
