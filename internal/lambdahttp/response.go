package lambdahttp

import (
	"bytes"
	"encoding/base64"
	"mime"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

// responseWriter buffers a handler's response so it can be returned to API
// Gateway in one piece. Like net/http, headers changed after WriteHeader are
// ignored, and a missing Content-Type is sniffed from the first write.
type responseWriter struct {
	header      http.Header
	sentHeader  http.Header
	status      int
	wroteHeader bool
	body        bytes.Buffer
}

func newResponseWriter() *responseWriter {
	return &responseWriter{header: make(http.Header)}
}

func (w *responseWriter) Header() http.Header {
	return w.header
}

func (w *responseWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.status = status
	w.sentHeader = w.header.Clone()
}

func (w *responseWriter) Write(p []byte) (int, error) {
	if !w.wroteHeader {
		if w.header.Get("Content-Type") == "" && len(p) > 0 {
			w.header.Set("Content-Type", http.DetectContentType(p))
		}
		w.WriteHeader(http.StatusOK)
	}
	return w.body.Write(p)
}

// toResponse converts the buffered response into an API Gateway HTTP API
// response.
func (w *responseWriter) toResponse() events.APIGatewayV2HTTPResponse {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	resp := events.APIGatewayV2HTTPResponse{
		StatusCode: w.status,
		Headers:    make(map[string]string, len(w.sentHeader)),
	}
	for key, values := range w.sentHeader {
		// Set-Cookie values can contain commas, so API Gateway takes them
		// separately instead of in a joined header.
		if http.CanonicalHeaderKey(key) == "Set-Cookie" {
			resp.Cookies = append(resp.Cookies, values...)
			continue
		}
		resp.Headers[key] = strings.Join(values, ", ")
	}

	body := w.body.Bytes()
	if len(body) == 0 || isTextual(w.sentHeader.Get("Content-Type")) {
		resp.Body = string(body)
	} else {
		resp.Body = base64.StdEncoding.EncodeToString(body)
		resp.IsBase64Encoded = true
	}
	return resp
}

func isTextual(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		mediaType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	}
	switch {
	case strings.HasPrefix(mediaType, "text/"),
		strings.HasSuffix(mediaType, "+json"),
		strings.HasSuffix(mediaType, "+xml"):
		return true
	}
	switch mediaType {
	case "application/json",
		"application/xml",
		"application/x-www-form-urlencoded",
		"application/javascript":
		return true
	}
	return false
}
