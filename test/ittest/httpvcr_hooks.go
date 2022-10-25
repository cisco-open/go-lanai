package ittest

import (
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

const (
	DefaultHost = "webservice"
)

type ValueSanitizer func(string) string

func RegExpValueSanitizer(regex, repl string) ValueSanitizer {
	pattern := regexp.MustCompile(regex)
	return func(str string) string {
		return pattern.ReplaceAllString(str, repl)
	}
}

func SubstituteValueSanitizer(repl string) ValueSanitizer {
	return func(_ string) string {
		return repl
	}
}

func DefaultValueSanitizer() ValueSanitizer {
	return SubstituteValueSanitizer("_hidden")
}

var (
	SensitiveHeaderSanitizer = map[string]ValueSanitizer{
		"Authorization": RegExpValueSanitizer("^(?P<prefix>Basic |Bearer |Digest ).*|.*", "${prefix}******"),
	}
	SensitiveQuerySanitizer = map[string]ValueSanitizer{
		"password": DefaultValueSanitizer(),
		"secret":   DefaultValueSanitizer(),
		"access_token": DefaultValueSanitizer(),
		"token": DefaultValueSanitizer(),
	}
	SensitiveBodySanitizer = map[string]ValueSanitizer{
		"access_token": DefaultValueSanitizer(),
		"secret":   DefaultValueSanitizer(),
	}
)

/************************
	Hooks
 ************************/

// InteractionIndexAwareHook inject interaction index into stored header:
// httpvcr store interaction's ID but doesn't expose it to cassette.MatchFunc,
// so we need to store it in request for request matchers to access
func InteractionIndexAwareHook() func(i *cassette.Interaction) error {
	return func(i *cassette.Interaction) error {
		i.Request.Headers.Set(xInteractionIndexHeader, strconv.Itoa(i.ID))
		return nil
	}
}

// SanitizingHook is a httpvcr hook that sanitize values in header, query, body (x-form-urlencoded/json)
func SanitizingHook() func(i *cassette.Interaction) error {
	return func(i *cassette.Interaction) error {
		i.Request.Headers = sanitizeHeaders(i.Request.Headers)
		i.Request.URL = sanitizeUrl(i.Request.URL)
		sanitizeRequestForm(&i.Request)
		i.Request.Body = sanitizeJsonBody(i.Request.Body)

		i.Response.Headers = sanitizeHeaders(i.Response.Headers)
		i.Response.Body = sanitizeJsonBody(i.Response.Body)
		return nil
	}
}

// HostIgnoringHook changes the host of request to a pre-defined constant, to avoid randomness
func HostIgnoringHook() func(i *cassette.Interaction) error {
	return func(i *cassette.Interaction) error {
		i.Request.URL = strings.Replace(i.Request.URL, i.Request.Host, DefaultHost, 1)
		i.Request.Host = DefaultHost
		return nil
	}
}

/************************
	helpers
 ************************/

func sanitizeValues(values map[string][]string, sanitizers map[string]ValueSanitizer) map[string][]string {
	for k := range values {
		sanitizer, ok := sanitizers[k]
		if !ok {
			continue
		}
		for i := range values[k] {
			values[k][i] = sanitizer(values[k][i])
		}
	}
	return values
}

func sanitizeHeaders(headers http.Header) http.Header {
	return sanitizeValues(headers, SensitiveHeaderSanitizer)
}

func sanitizeUrl(raw string) string {
	parsed, e := url.Parse(raw)
	if e != nil {
		return raw
	}
	var queries url.Values = sanitizeValues(parsed.Query(), SensitiveQuerySanitizer)
	parsed.RawQuery = queries.Encode()
	return parsed.String()
}

func sanitizeRequestForm(req *cassette.Request) {
	req.Form = sanitizeValues(req.Form, SensitiveQuerySanitizer)
	req.Body = req.Form.Encode()
}

func sanitizeJsonBody(body string) string{
	// TODO
	return body
}
