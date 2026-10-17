package lambdahttp

import (
	"context"
	"encoding/base64"
	"net/http"
	"reflect"
	"testing"
)

func TestResponseTranslation(t *testing.T) {
	binary := []byte("\x89PNG\r\n\x1a\n\x00\xff")

	cases := []struct {
		name    string
		handler http.HandlerFunc
		status  int
		headers map[string]string
		cookies []string
		body    string
		base64  bool
	}{
		{
			name:    "default status",
			handler: func(w http.ResponseWriter, r *http.Request) {},
			status:  http.StatusOK,
			headers: map[string]string{},
		},
		{
			name: "default status on write",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				w.Write([]byte("hi"))
			},
			status:  http.StatusOK,
			headers: map[string]string{"Content-Type": "text/plain"},
			body:    "hi",
		},
		{
			name: "json",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.WriteHeader(http.StatusAccepted)
				w.Write([]byte(`{"message":"Pong from donkeyvpn"}`))
			},
			status:  http.StatusAccepted,
			headers: map[string]string{"Content-Type": "application/json; charset=UTF-8"},
			body:    `{"message":"Pong from donkeyvpn"}`,
		},
		{
			name: "csv",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/csv")
				w.WriteHeader(http.StatusAccepted)
				w.Write([]byte("10.0.0.2,pubkey1\n10.0.0.3,pubkey2\n"))
			},
			status:  http.StatusAccepted,
			headers: map[string]string{"Content-Type": "text/csv"},
			body:    "10.0.0.2,pubkey1\n10.0.0.3,pubkey2\n",
		},
		{
			name: "binary content type",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "image/png")
				w.Write(binary)
			},
			status:  http.StatusOK,
			headers: map[string]string{"Content-Type": "image/png"},
			body:    base64.StdEncoding.EncodeToString(binary),
			base64:  true,
		},
		{
			name: "missing content type is sniffed",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write(binary)
			},
			status:  http.StatusOK,
			headers: map[string]string{"Content-Type": "image/png"},
			body:    base64.StdEncoding.EncodeToString(binary),
			base64:  true,
		},
		{
			name: "two values for one header",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Vary", "Accept")
				w.Header().Add("Vary", "Accept-Encoding")
				w.WriteHeader(http.StatusNoContent)
			},
			status:  http.StatusNoContent,
			headers: map[string]string{"Vary": "Accept, Accept-Encoding"},
		},
		{
			name: "set-cookie goes to cookies",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Set-Cookie", "a=1; Expires=Wed, 21 Oct 2026 07:28:00 GMT")
				w.Header().Add("Set-Cookie", "b=2")
			},
			status:  http.StatusOK,
			headers: map[string]string{},
			cookies: []string{"a=1; Expires=Wed, 21 Oct 2026 07:28:00 GMT", "b=2"},
		},
		{
			name: "headers after WriteHeader are ignored",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusCreated)
				w.Header().Set("X-Late", "1")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("ok"))
			},
			status:  http.StatusCreated,
			headers: map[string]string{"Content-Type": "text/plain"},
			body:    "ok",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := Handler(tc.handler)(context.Background(), newEvent("GET", "/"))
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tc.status {
				t.Errorf("expected status %d, got %d", tc.status, resp.StatusCode)
			}
			if !reflect.DeepEqual(resp.Headers, tc.headers) {
				t.Errorf("expected headers %v, got %v", tc.headers, resp.Headers)
			}
			if !reflect.DeepEqual(resp.Cookies, tc.cookies) {
				t.Errorf("expected cookies %v, got %v", tc.cookies, resp.Cookies)
			}
			if resp.Body != tc.body || resp.IsBase64Encoded != tc.base64 {
				t.Errorf("expected body %q (base64=%v), got %q (base64=%v)", tc.body, tc.base64, resp.Body, resp.IsBase64Encoded)
			}
		})
	}
}

func TestIsTextual(t *testing.T) {
	cases := map[string]bool{
		"text/plain; charset=UTF-8":         true,
		"text/csv":                          true,
		"application/json":                  true,
		"application/problem+json":          true,
		"application/xml":                   true,
		"application/atom+xml":              true,
		"application/x-www-form-urlencoded": true,
		"application/javascript":            true,
		"Application/JSON":                  true,
		"application/octet-stream":          false,
		"image/png":                         false,
		"":                                  false,
	}
	for contentType, want := range cases {
		if got := isTextual(contentType); got != want {
			t.Errorf("isTextual(%q) = %v, want %v", contentType, got, want)
		}
	}
}

func TestHandlerInvalidEvent(t *testing.T) {
	e := newEvent("POST", "/")
	e.Body = "not base64!"
	e.IsBase64Encoded = true
	called := false
	_, err := Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }))(context.Background(), e)
	if err == nil || called {
		t.Errorf("expected an error and no handler call, got err=%v called=%v", err, called)
	}
}
