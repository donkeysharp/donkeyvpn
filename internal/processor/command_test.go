package processor

import (
	"errors"
	"testing"

	"github.com/donkeysharp/donkeyvpn/internal/telegram"
)

type recordingProcessor struct {
	calls int
}

func (p *recordingProcessor) Process(args []string, update *telegram.Update) error {
	p.calls++
	return nil
}

func newTestUpdate(chatId telegram.ChatId, text string) *telegram.Update {
	return &telegram.Update{
		Message: &telegram.Message{
			Chat: &telegram.Chat{ChatId: chatId},
			Text: text,
		},
	}
}

type gate struct {
	cmd          *CommandProcessor
	list         *recordingProcessor
	fallback     *recordingProcessor
	unauthorized *recordingProcessor
}

func newGate(allowed []int64) gate {
	g := gate{
		list:         &recordingProcessor{},
		fallback:     &recordingProcessor{},
		unauthorized: &recordingProcessor{},
	}
	loader := func() ([]int64, error) { return allowed, nil }
	g.cmd = NewCommandProcessor(NewAuthorizer(loader), g.unauthorized)
	g.cmd.Register("/list", g.list)
	g.cmd.RegisterFallback(g.fallback)
	return g
}

func TestAllowedChatRunsCommand(t *testing.T) {
	g := newGate([]int64{12345})
	g.cmd.Process(newTestUpdate(12345, "/list vpn"))
	if g.list.calls != 1 || g.unauthorized.calls != 0 {
		t.Errorf("expected command to run, list=%d unauthorized=%d", g.list.calls, g.unauthorized.calls)
	}
}

func TestAllowedGroupRunsCommand(t *testing.T) {
	g := newGate([]int64{-100123})
	g.cmd.Process(newTestUpdate(-100123, "/list vpn"))
	if g.list.calls != 1 || g.unauthorized.calls != 0 {
		t.Errorf("expected command to run, list=%d unauthorized=%d", g.list.calls, g.unauthorized.calls)
	}
}

func TestUnknownChatOnlyGetsUnauthorized(t *testing.T) {
	g := newGate([]int64{12345})
	g.cmd.Process(newTestUpdate(99999, "/list vpn"))
	g.cmd.Process(newTestUpdate(99999, "/help"))
	if g.list.calls != 0 || g.fallback.calls != 0 {
		t.Errorf("expected no command to run, list=%d fallback=%d", g.list.calls, g.fallback.calls)
	}
	if g.unauthorized.calls != 2 {
		t.Errorf("expected one not-authorized reply per message, got %d", g.unauthorized.calls)
	}
}

func TestEmptyAllowlistDenies(t *testing.T) {
	g := newGate(nil)
	g.cmd.Process(newTestUpdate(12345, "/list vpn"))
	if g.list.calls != 0 || g.unauthorized.calls != 1 {
		t.Errorf("expected deny, list=%d unauthorized=%d", g.list.calls, g.unauthorized.calls)
	}
}

func TestAuthorizerLoadsOnceAndCaches(t *testing.T) {
	calls := 0
	a := NewAuthorizer(func() ([]int64, error) {
		calls++
		return []int64{12345}, nil
	})
	a.IsAllowed(12345)
	a.IsAllowed(99999)
	if calls != 1 {
		t.Errorf("expected the allowlist to be loaded once, got %d loads", calls)
	}
}

func TestAuthorizerLoadErrorDeniesAndRetries(t *testing.T) {
	calls := 0
	a := NewAuthorizer(func() ([]int64, error) {
		calls++
		if calls == 1 {
			return nil, errors.New("ssm unavailable")
		}
		return []int64{12345}, nil
	})
	if a.IsAllowed(12345) {
		t.Error("expected deny while the allowlist can't be loaded")
	}
	if !a.IsAllowed(12345) {
		t.Error("expected allow after a successful retry")
	}
}

func TestNilLoaderDenies(t *testing.T) {
	if NewAuthorizer(nil).IsAllowed(12345) {
		t.Error("authorizer without a loader must deny")
	}
}

func TestNilAuthorizerDenies(t *testing.T) {
	var a *Authorizer
	if a.IsAllowed(12345) {
		t.Error("nil authorizer must deny")
	}
}

func TestUnauthorizedMessage(t *testing.T) {
	expected := "⛔ You are not authorized to use this bot. Your chat id is `99999`."
	if got := unauthorizedMessage(99999); got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
	expected = "⛔ You are not authorized to use this bot. Your chat id is `-100123`."
	if got := unauthorizedMessage(-100123); got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}
