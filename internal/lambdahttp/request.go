// Package lambdahttp serves an http.Handler from AWS Lambda behind an
// API Gateway HTTP API (payload format v2).
package lambdahttp

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

// toHTTPRequest converts an API Gateway HTTP API event into a server-side
// *http.Request, as net/http would build it for the same request.
func toHTTPRequest(ctx context.Context, event events.APIGatewayV2HTTPRequest) (*http.Request, error) {
	body := []byte(event.Body)
	if event.IsBase64Encoded {
		decoded, err := base64.StdEncoding.DecodeString(event.Body)
		if err != nil {
			return nil, fmt.Errorf("decoding base64 body: %w", err)
		}
		body = decoded
	}

	path := event.RawPath
	if path == "" {
		path = event.RequestContext.HTTP.Path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	// RawQueryString keeps repeated keys and their order; QueryStringParameters
	// comma-joins them.
	requestURI := path
	if event.RawQueryString != "" {
		requestURI += "?" + event.RawQueryString
	}
	u, err := url.ParseRequestURI(requestURI)
	if err != nil {
		return nil, fmt.Errorf("parsing request URI %q: %w", requestURI, err)
	}

	header := make(http.Header, len(event.Headers)+1)
	host := event.RequestContext.DomainName
	for key, value := range event.Headers {
		// Like net/http, the host lives in req.Host rather than in the header map.
		if strings.EqualFold(key, "host") {
			if host == "" {
				host = value
			}
			continue
		}
		header.Add(key, value)
	}
	if len(event.Cookies) > 0 {
		header.Set("Cookie", strings.Join(event.Cookies, "; "))
	}

	req := &http.Request{
		Method:        strings.ToUpper(event.RequestContext.HTTP.Method),
		URL:           u,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        header,
		Body:          http.NoBody,
		ContentLength: int64(len(body)),
		Host:          host,
		RemoteAddr:    event.RequestContext.HTTP.SourceIP,
		RequestURI:    requestURI,
	}
	if len(body) > 0 {
		req.Body = io.NopCloser(bytes.NewReader(body))
	}
	return req.WithContext(ctx), nil
}
