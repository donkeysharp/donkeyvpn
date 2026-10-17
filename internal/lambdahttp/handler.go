package lambdahttp

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

// Handler adapts h into a Lambda handler for API Gateway HTTP API (payload
// v2) events, for use with lambda.Start.
func Handler(h http.Handler) func(context.Context, events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return func(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		req, err := toHTTPRequest(ctx, event)
		if err != nil {
			return events.APIGatewayV2HTTPResponse{}, err
		}
		w := newResponseWriter()
		h.ServeHTTP(w, req)
		return w.toResponse(), nil
	}
}
