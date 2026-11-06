package app

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/donkeysharp/donkeyvpn/internal/models"
	"github.com/donkeysharp/donkeyvpn/internal/service"
	"github.com/donkeysharp/donkeyvpn/internal/telegram"
	"github.com/donkeysharp/donkeyvpn/internal/testutil"
)

func newTestNotifier(t *testing.T, sender *testutil.RecordingSender, instances ...*models.VPNInstance) *DonkeyVPNNotifierApp {
	t.Helper()
	now := time.Now()
	table := testutil.NewFakeTable()
	ec2 := &testutil.FakeInstanceDescriber{Instances: map[string]ec2types.Instance{}}
	for _, instance := range instances {
		if _, err := table.CreateRecord(instance); err != nil {
			t.Fatal(err)
		}
		// instance ids ending in "-new" were launched recently, all others two hours ago
		launched := now.Add(-2 * time.Hour)
		if strings.HasSuffix(instance.InstanceId, "-new") {
			launched = now.Add(-10 * time.Minute)
		}
		ec2.Instances[instance.InstanceId] = testutil.NewEC2Instance(instance.InstanceId, ec2types.InstanceStateNameRunning, launched)
	}
	return &DonkeyVPNNotifierApp{
		VPNService:     service.NewVPNService(nil, table, ec2),
		TelegramClient: sender,
	}
}

func sortedByChat(messages []testutil.SentMessage) []testutil.SentMessage {
	slices.SortFunc(messages, func(a, b testutil.SentMessage) int { return int(a.ChatId - b.ChatId) })
	return messages
}

func TestCheckInstancesSendsOneMessagePerChat(t *testing.T) {
	sender := &testutil.RecordingSender{}
	notifier := newTestNotifier(t, sender,
		models.NewVPNInstance("vpn001", "", "", "success", "i-1", "42"),
		models.NewVPNInstance("vpn002", "", "", "success", "i-2", "42"),
		models.NewVPNInstance("vpn003", "", "", "success", "i-3", "-100123"),
		models.NewVPNInstance("vpn004", "", "", "success", "i-4-new", "7"),
	)

	if err := notifier.CheckInstances(); err != nil {
		t.Fatal(err)
	}

	// Current behavior: every chat that owns an old instance gets the same
	// message, which lists the old instances of all chats.
	want := "⚠️ Warning! The next instances have been running for *more than 1 hour:*\n\n" +
		"🟢 Instance ID: `vpn001`\n" +
		"🟢 Instance ID: `vpn002`\n" +
		"🟢 Instance ID: `vpn003`\n" +
		"----\n" +
		"In case your are not using them, you can delete them and create new ones later 🤑"
	got := sortedByChat(sender.Messages())
	if len(got) != 2 {
		t.Fatalf("expected one message for each of the 2 chats, got %d: %v", len(got), got)
	}
	if got[0].ChatId != telegram.ChatId(-100123) || got[1].ChatId != telegram.ChatId(42) {
		t.Errorf("expected messages to chats -100123 and 42, got %d and %d", got[0].ChatId, got[1].ChatId)
	}
	for _, m := range got {
		if m.Text != want {
			t.Errorf("chat %d: expected\n%q\ngot\n%q", m.ChatId, want, m.Text)
		}
	}
}

func TestCheckInstancesWithNoOldInstancesSendsNothing(t *testing.T) {
	sender := &testutil.RecordingSender{}
	notifier := newTestNotifier(t, sender,
		models.NewVPNInstance("vpn001", "", "", "success", "i-1-new", "42"),
		models.NewVPNInstance("vpn002", "", "", service.STATUS_PENDING, "pending", "42"),
	)

	if err := notifier.CheckInstances(); err != nil {
		t.Fatal(err)
	}
	if got := sender.Messages(); len(got) != 0 {
		t.Errorf("expected no messages, got %v", got)
	}
}

func TestCheckInstancesKeepsGoingWhenSendFails(t *testing.T) {
	sender := &testutil.RecordingSender{Err: errors.New("telegram is down")}
	notifier := newTestNotifier(t, sender,
		models.NewVPNInstance("vpn001", "", "", "success", "i-1", "42"),
		models.NewVPNInstance("vpn002", "", "", "success", "i-2", "7"),
	)

	if err := notifier.CheckInstances(); err != nil {
		t.Errorf("expected send failures to be logged, not returned, got %v", err)
	}
	if got := sender.Messages(); len(got) != 2 {
		t.Errorf("expected a send attempt for each chat, got %d", len(got))
	}
}
