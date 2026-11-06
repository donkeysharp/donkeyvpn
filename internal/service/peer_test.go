package service_test

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/donkeysharp/donkeyvpn/internal/models"
	"github.com/donkeysharp/donkeyvpn/internal/service"
	"github.com/donkeysharp/donkeyvpn/internal/testutil"
)

const testCidr = "10.0.0.0/24"

var validKey = base64.StdEncoding.EncodeToString([]byte(strings.Repeat("k", 32)))

func TestPeerServiceCreate(t *testing.T) {
	cases := []struct {
		name      string
		ip        string
		key       string
		wantErr   error
		wantSaved bool
	}{
		{name: "success", ip: "10.0.0.2", key: validKey, wantSaved: true},
		{name: "key is not base64", ip: "10.0.0.2", key: "not a key!", wantErr: service.ErrInvalidWireguardKey},
		{name: "key is not 32 bytes", ip: "10.0.0.2", key: base64.StdEncoding.EncodeToString([]byte("short")), wantErr: service.ErrInvalidWireguardKey},
		{name: "empty key", ip: "10.0.0.2", key: "", wantErr: service.ErrInvalidWireguardKey},
		{name: "ip outside range", ip: "10.0.1.2", key: validKey, wantErr: service.ErrInvalidIPAddress},
		{name: "ip is not an ip", ip: "vpn001", key: validKey, wantErr: service.ErrInvalidIPAddress},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			table := testutil.NewFakeTable()
			svc := service.NewWireguardPeerService(table, testCidr)

			created, err := svc.Create(models.NewWireguardPeer(tc.ip, tc.key, "alice"))
			if err != tc.wantErr || created != tc.wantSaved {
				t.Fatalf("expected created=%v err=%v, got created=%v err=%v", tc.wantSaved, tc.wantErr, created, err)
			}
			if !tc.wantSaved {
				if table.Len() != 0 {
					t.Errorf("expected no record to be written, got %d", table.Len())
				}
				return
			}
			peer, err := svc.Get(tc.ip)
			if err != nil {
				t.Fatal(err)
			}
			if want := *models.NewWireguardPeer(tc.ip, tc.key, "alice"); *peer != want {
				t.Errorf("expected stored peer %v, got %v", want, peer)
			}
		})
	}
}

func TestPeerServiceCidrRange(t *testing.T) {
	if got := service.NewWireguardPeerService(testutil.NewFakeTable(), testCidr).CidrRange(); got != testCidr {
		t.Errorf("expected %q, got %q", testCidr, got)
	}
}
