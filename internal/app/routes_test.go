package app

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/donkeysharp/donkeyvpn/internal/aws"
	"github.com/donkeysharp/donkeyvpn/internal/handler"
	"github.com/donkeysharp/donkeyvpn/internal/processor"
	"github.com/donkeysharp/donkeyvpn/internal/service"
	"github.com/donkeysharp/donkeyvpn/internal/telegram"
	"github.com/labstack/echo/v4"
)

const (
	instancesTable = "test-instances"
	peersTable     = "test-peers"
)

// fakeDynamoDB answers the DynamoDB JSON API calls the handlers make (Scan,
// GetItem, UpdateItem) with canned items, and records each call.
type fakeDynamoDB struct {
	mu    sync.Mutex
	calls []fakeCall
}

type fakeCall struct {
	op   string
	body map[string]any
}

var fakeItems = map[string][]map[string]any{
	instancesTable: {
		{"Id": map[string]string{"S": "vpn002"}, "Status": map[string]string{"S": "pending"}},
	},
	peersTable: {
		{"IPAddress": map[string]string{"S": "10.0.0.2"}, "PublicKey": map[string]string{"S": "pubkey1"}},
		{"IPAddress": map[string]string{"S": "10.0.0.3"}, "PublicKey": map[string]string{"S": "pubkey2"}},
	},
}

func (f *fakeDynamoDB) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	op := strings.TrimPrefix(r.Header.Get("X-Amz-Target"), "DynamoDB_20120810.")
	var body map[string]any
	raw, _ := io.ReadAll(r.Body)
	json.Unmarshal(raw, &body)
	f.mu.Lock()
	f.calls = append(f.calls, fakeCall{op: op, body: body})
	f.mu.Unlock()

	table, _ := body["TableName"].(string)
	items := fakeItems[table]
	var resp map[string]any
	switch op {
	case "Scan":
		resp = map[string]any{"Items": items, "Count": len(items), "ScannedCount": len(items)}
	case "GetItem":
		resp = map[string]any{}
		key, _ := json.Marshal(body["Key"])
		for _, item := range items {
			if id, _ := json.Marshal(map[string]any{"Id": item["Id"]}); string(id) == string(key) {
				resp["Item"] = item
			}
		}
	case "UpdateItem":
		resp = map[string]any{}
	default:
		w.WriteHeader(http.StatusBadRequest)
		resp = map[string]any{"__type": "com.amazon.coral.service#UnknownOperationException"}
	}
	w.Header().Set("Content-Type", "application/x-amz-json-1.0")
	json.NewEncoder(w).Encode(resp)
}

func (f *fakeDynamoDB) find(op string) []fakeCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	var found []fakeCall
	for _, c := range f.calls {
		if c.op == op {
			found = append(found, c)
		}
	}
	return found
}

type recordingProcessor struct {
	mu   sync.Mutex
	args [][]string
}

func (p *recordingProcessor) Process(args []string, update *telegram.Update) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.args = append(p.args, args)
	return nil
}

// newTestApp wires the real routes, middleware, handlers and services. AWS
// clients talk to fakeDynamoDB, so nothing reaches real AWS.
func newTestApp(t *testing.T) (*DonkeyVPNApplication, *fakeDynamoDB, *recordingProcessor) {
	t.Helper()
	fake := &fakeDynamoDB{}
	server := httptest.NewServer(fake)
	t.Cleanup(server.Close)

	t.Setenv("AWS_ENDPOINT_URL", server.URL)
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "fake")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "fake")
	t.Setenv("AWS_SESSION_TOKEN", "")
	t.Setenv("AWS_PROFILE", "")
	t.Setenv("AWS_CONFIG_FILE", "/dev/null")
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", "/dev/null")
	t.Setenv("AWS_EC2_METADATA_DISABLED", "true")

	ctx := context.Background()
	instances, err := aws.NewDynamoDB(ctx, instancesTable)
	if err != nil {
		t.Fatal(err)
	}
	peers, err := aws.NewDynamoDB(ctx, peersTable)
	if err != nil {
		t.Fatal(err)
	}

	list := &recordingProcessor{}
	cmd := processor.NewCommandProcessor(
		processor.NewAuthorizer(func() ([]int64, error) { return []int64{42}, nil }),
		&recordingProcessor{},
	)
	cmd.Register("/list", list)
	cmd.RegisterFallback(&recordingProcessor{})

	app := &DonkeyVPNApplication{
		e:              echo.New(),
		webhookSecret:  configuredSecret,
		webhookHandler: &handler.WebhookHandler{WebhookSecret: configuredSecret, CommandProcessor: cmd},
		vpnHandler:     &handler.VPNHandler{WebhookSecret: configuredSecret, VPNSvc: service.NewVPNService(nil, instances, nil)},
		peerHandler:    &handler.PeerHandler{WebhookSecret: configuredSecret, PeersTable: peers},
	}
	app.registerRoutes()
	return app, fake, list
}

func TestRoutes(t *testing.T) {
	cases := []struct {
		name        string
		method      string
		path        string
		headers     map[string]string
		body        string
		status      int
		contentType string
		check       func(t *testing.T, body string, fake *fakeDynamoDB, list *recordingProcessor)
	}{
		{
			name:        "ping",
			method:      http.MethodGet,
			path:        "/donkeyvpn/ping",
			status:      http.StatusAccepted,
			contentType: "application/json",
			check: func(t *testing.T, body string, _ *fakeDynamoDB, _ *recordingProcessor) {
				var got map[string]string
				if err := json.Unmarshal([]byte(body), &got); err != nil || got["message"] != "Pong from donkeyvpn" {
					t.Errorf("unexpected body %q", body)
				}
			},
		},
		{
			name:   "webhook",
			method: http.MethodPost,
			path:   "/v1/api/telegram/donkeyvpn/webhook",
			headers: map[string]string{
				"content-type":                    "application/json",
				"x-telegram-bot-api-secret-token": configuredSecret,
			},
			body: `{"update_id":1,"message":{"message_id":2,"date":1700000000,"text":"/list vpn",` +
				`"from":{"id":42},"chat":{"id":42,"type":"private"}}}`,
			status: http.StatusOK,
			check: func(t *testing.T, _ string, _ *fakeDynamoDB, list *recordingProcessor) {
				if len(list.args) != 1 || strings.Join(list.args[0], " ") != "vpn" {
					t.Errorf("expected /list to run once with [vpn], got %v", list.args)
				}
			},
		},
		{
			name:        "pending vpn",
			method:      http.MethodGet,
			path:        "/v1/api/vpn/pending",
			headers:     map[string]string{"x-api-key": configuredSecret},
			status:      http.StatusAccepted,
			contentType: "application/json",
			check: func(t *testing.T, body string, fake *fakeDynamoDB, _ *recordingProcessor) {
				// The boot script reads `.[0].Id`.
				var got []map[string]any
				if err := json.Unmarshal([]byte(body), &got); err != nil || len(got) != 1 || got[0]["Id"] != "vpn002" {
					t.Errorf("unexpected body %q", body)
				}
				scans := fake.find("Scan")
				if len(scans) != 1 || scans[0].body["FilterExpression"] != "#status = :status" {
					t.Errorf("expected one filtered Scan, got %v", scans)
				}
			},
		},
		{
			name:        "notify",
			method:      http.MethodPost,
			path:        "/v1/api/vpn/notify/vpn002",
			headers:     map[string]string{"x-api-key": configuredSecret, "content-type": "application/json"},
			body:        `{"hostname":"vpn002.example.com","status":"success","instanceId":"i-0abc","port":"51820"}`,
			status:      http.StatusAccepted,
			contentType: "application/json",
			check: func(t *testing.T, _ string, fake *fakeDynamoDB, _ *recordingProcessor) {
				updates := fake.find("UpdateItem")
				if len(updates) != 1 {
					t.Fatalf("expected one UpdateItem, got %d", len(updates))
				}
				got, _ := json.Marshal(updates[0].body)
				for _, want := range []string{`"vpn002"`, `"vpn002.example.com"`, `"success"`, `"i-0abc"`, `"51820"`} {
					if !strings.Contains(string(got), want) {
						t.Errorf("expected UpdateItem to contain %s, got %s", want, got)
					}
				}
			},
		},
		{
			name:        "notify unknown vpn",
			method:      http.MethodPost,
			path:        "/v1/api/vpn/notify/vpn005",
			headers:     map[string]string{"x-api-key": configuredSecret, "content-type": "application/json"},
			body:        `{"hostname":"vpn005.example.com","status":"success","instanceId":"i-0def","port":"51820"}`,
			status:      http.StatusNotFound,
			contentType: "application/json",
			check: func(t *testing.T, _ string, fake *fakeDynamoDB, _ *recordingProcessor) {
				if n := len(fake.find("UpdateItem")); n != 0 {
					t.Errorf("expected no UpdateItem, got %d", n)
				}
			},
		},
		{
			name:        "peer csv",
			method:      http.MethodGet,
			path:        "/v1/api/peer",
			headers:     map[string]string{"x-api-key": configuredSecret},
			status:      http.StatusAccepted,
			contentType: "text/csv",
			check: func(t *testing.T, body string, _ *fakeDynamoDB, _ *recordingProcessor) {
				if body != "10.0.0.2,pubkey1\n10.0.0.3,pubkey2\n" {
					t.Errorf("unexpected body %q", body)
				}
			},
		},
		{
			name:   "401 without a secret",
			method: http.MethodGet,
			path:   "/v1/api/peer",
			status: http.StatusUnauthorized,
			check: func(t *testing.T, _ string, fake *fakeDynamoDB, _ *recordingProcessor) {
				if n := len(fake.find("Scan")); n != 0 {
					t.Errorf("expected the handler not to run, got %d Scan calls", n)
				}
			},
		},
		{
			name:   "404 unknown path",
			method: http.MethodGet,
			path:   "/does-not-exist",
			status: http.StatusNotFound,
		},
		{
			name:   "405 wrong method",
			method: http.MethodPost,
			path:   "/donkeyvpn/ping",
			status: http.StatusMethodNotAllowed,
		},
		{
			// Echo v4 answers 404 here, not 405: Group.Use registers a
			// catch-all route for the group that matches first.
			name:    "wrong method on an authenticated route",
			method:  http.MethodGet,
			path:    "/v1/api/vpn/notify/vpn002",
			headers: map[string]string{"x-api-key": configuredSecret},
			status:  http.StatusNotFound,
			check: func(t *testing.T, _ string, fake *fakeDynamoDB, _ *recordingProcessor) {
				if len(fake.calls) != 0 {
					t.Errorf("expected no handler to run, got %d DynamoDB calls", len(fake.calls))
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app, fake, list := newTestApp(t)

			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			rec := httptest.NewRecorder()
			app.e.ServeHTTP(rec, req)

			if rec.Code != tc.status {
				t.Errorf("expected status %d, got %d (body %q)", tc.status, rec.Code, rec.Body.String())
			}
			if tc.contentType != "" && !strings.HasPrefix(rec.Header().Get("Content-Type"), tc.contentType) {
				t.Errorf("expected content type %q, got %q", tc.contentType, rec.Header().Get("Content-Type"))
			}
			if tc.check != nil {
				tc.check(t, rec.Body.String(), fake, list)
			}
		})
	}
}
