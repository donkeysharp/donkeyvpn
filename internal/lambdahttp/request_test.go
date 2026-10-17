package lambdahttp

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func newEvent(method, rawPath string) events.APIGatewayV2HTTPRequest {
	return events.APIGatewayV2HTTPRequest{
		RawPath: rawPath,
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			DomainName: "abc123.execute-api.us-east-1.amazonaws.com",
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method:   method,
				Path:     rawPath,
				SourceIP: "203.0.113.7",
			},
		},
	}
}

func TestToHTTPRequest(t *testing.T) {
	cases := []struct {
		name  string
		event func() events.APIGatewayV2HTTPRequest
		check func(t *testing.T, req *http.Request)
	}{
		{
			name:  "method",
			event: func() events.APIGatewayV2HTTPRequest { return newEvent("post", "/v1/api/peer") },
			check: func(t *testing.T, req *http.Request) {
				if req.Method != http.MethodPost {
					t.Errorf("expected method POST, got %q", req.Method)
				}
			},
		},
		{
			name:  "raw path keeps escapes",
			event: func() events.APIGatewayV2HTTPRequest { return newEvent("GET", "/v1/api/vpn/notify/vpn%2F002") },
			check: func(t *testing.T, req *http.Request) {
				if got := req.URL.EscapedPath(); got != "/v1/api/vpn/notify/vpn%2F002" {
					t.Errorf("expected escaped path to be preserved, got %q", got)
				}
				if got := req.URL.Path; got != "/v1/api/vpn/notify/vpn/002" {
					t.Errorf("expected decoded path, got %q", got)
				}
			},
		},
		{
			name: "path falls back to request context",
			event: func() events.APIGatewayV2HTTPRequest {
				e := newEvent("GET", "/donkeyvpn/ping")
				e.RawPath = ""
				return e
			},
			check: func(t *testing.T, req *http.Request) {
				if req.URL.Path != "/donkeyvpn/ping" {
					t.Errorf("expected /donkeyvpn/ping, got %q", req.URL.Path)
				}
			},
		},
		{
			name: "repeated query keys",
			event: func() events.APIGatewayV2HTTPRequest {
				e := newEvent("GET", "/v1/api/peer")
				e.RawQueryString = "a=1&a=2&b=x%20y"
				e.QueryStringParameters = map[string]string{"a": "1,2", "b": "x y"}
				return e
			},
			check: func(t *testing.T, req *http.Request) {
				q := req.URL.Query()
				if !reflect.DeepEqual(q["a"], []string{"1", "2"}) {
					t.Errorf("expected a=[1 2], got %v", q["a"])
				}
				if q.Get("b") != "x y" {
					t.Errorf("expected b=%q, got %q", "x y", q.Get("b"))
				}
				if req.RequestURI != "/v1/api/peer?a=1&a=2&b=x%20y" {
					t.Errorf("unexpected RequestURI %q", req.RequestURI)
				}
			},
		},
		{
			name: "headers are case-insensitive",
			event: func() events.APIGatewayV2HTTPRequest {
				e := newEvent("POST", "/v1/api/telegram/donkeyvpn/webhook")
				e.Headers = map[string]string{
					"x-telegram-bot-api-secret-token": "s3cret",
					"accept":                          "text/html, application/json",
				}
				return e
			},
			check: func(t *testing.T, req *http.Request) {
				if got := req.Header.Get("X-Telegram-Bot-Api-Secret-Token"); got != "s3cret" {
					t.Errorf("expected secret header, got %q", got)
				}
				if got := req.Header.Values("Accept"); !reflect.DeepEqual(got, []string{"text/html, application/json"}) {
					t.Errorf("expected comma-joined header kept as one value, got %v", got)
				}
			},
		},
		{
			name: "cookies",
			event: func() events.APIGatewayV2HTTPRequest {
				e := newEvent("GET", "/donkeyvpn/ping")
				e.Cookies = []string{"a=1", "b=2"}
				return e
			},
			check: func(t *testing.T, req *http.Request) {
				if got := req.Header.Values("Cookie"); !reflect.DeepEqual(got, []string{"a=1; b=2"}) {
					t.Errorf("expected one joined Cookie header, got %v", got)
				}
				if c, err := req.Cookie("b"); err != nil || c.Value != "2" {
					t.Errorf("expected cookie b=2, got %v (%v)", c, err)
				}
			},
		},
		{
			name: "plain body",
			event: func() events.APIGatewayV2HTTPRequest {
				e := newEvent("POST", "/v1/api/vpn/notify/vpn002")
				e.Body = `{"status":"success"}`
				return e
			},
			check: func(t *testing.T, req *http.Request) {
				assertBody(t, req, []byte(`{"status":"success"}`))
			},
		},
		{
			name: "base64 body",
			event: func() events.APIGatewayV2HTTPRequest {
				e := newEvent("POST", "/v1/api/vpn/notify/vpn002")
				e.Body = base64.StdEncoding.EncodeToString([]byte{0x00, 0xff, 'h', 'i'})
				e.IsBase64Encoded = true
				return e
			},
			check: func(t *testing.T, req *http.Request) {
				assertBody(t, req, []byte{0x00, 0xff, 'h', 'i'})
			},
		},
		{
			name:  "empty body",
			event: func() events.APIGatewayV2HTTPRequest { return newEvent("GET", "/v1/api/peer") },
			check: func(t *testing.T, req *http.Request) {
				if req.Body != http.NoBody {
					t.Errorf("expected http.NoBody for an empty body")
				}
				assertBody(t, req, []byte{})
			},
		},
		{
			name:  "host from domain name",
			event: func() events.APIGatewayV2HTTPRequest { return newEvent("GET", "/donkeyvpn/ping") },
			check: func(t *testing.T, req *http.Request) {
				if req.Host != "abc123.execute-api.us-east-1.amazonaws.com" {
					t.Errorf("unexpected host %q", req.Host)
				}
			},
		},
		{
			name: "host header is not copied into the header map",
			event: func() events.APIGatewayV2HTTPRequest {
				e := newEvent("GET", "/donkeyvpn/ping")
				e.RequestContext.DomainName = ""
				e.Headers = map[string]string{"host": "api.example.com"}
				return e
			},
			check: func(t *testing.T, req *http.Request) {
				if req.Host != "api.example.com" {
					t.Errorf("expected host from header, got %q", req.Host)
				}
				if _, ok := req.Header["Host"]; ok {
					t.Errorf("expected no Host entry in the header map")
				}
			},
		},
		{
			name:  "source ip",
			event: func() events.APIGatewayV2HTTPRequest { return newEvent("GET", "/donkeyvpn/ping") },
			check: func(t *testing.T, req *http.Request) {
				if req.RemoteAddr != "203.0.113.7" {
					t.Errorf("unexpected remote addr %q", req.RemoteAddr)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := toHTTPRequest(context.Background(), tc.event())
			if err != nil {
				t.Fatal(err)
			}
			tc.check(t, req)
		})
	}
}

func TestToHTTPRequestContext(t *testing.T) {
	type key struct{}
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), key{}, "lambda"))

	req, err := toHTTPRequest(ctx, newEvent("GET", "/donkeyvpn/ping"))
	if err != nil {
		t.Fatal(err)
	}
	if req.Context().Value(key{}) != "lambda" {
		t.Errorf("expected the Lambda context to be attached")
	}
	cancel()
	if req.Context().Err() == nil {
		t.Errorf("expected cancellation to propagate")
	}
}

func TestToHTTPRequestInvalidBase64(t *testing.T) {
	e := newEvent("POST", "/v1/api/peer")
	e.Body = "not base64!"
	e.IsBase64Encoded = true
	if _, err := toHTTPRequest(context.Background(), e); err == nil {
		t.Errorf("expected an error for an invalid base64 body")
	}
}

func assertBody(t *testing.T, req *http.Request, want []byte) {
	t.Helper()
	got, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Errorf("expected body %q, got %q", want, got)
	}
	if req.ContentLength != int64(len(want)) {
		t.Errorf("expected content length %d, got %d", len(want), req.ContentLength)
	}
}
