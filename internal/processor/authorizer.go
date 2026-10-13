package processor

import (
	"fmt"
	"sync"

	"github.com/donkeysharp/donkeyvpn/internal/telegram"
	"github.com/labstack/gommon/log"
)

type AllowedChatIdsLoader func() ([]int64, error)

// The allowlist is loaded on first use and kept for the life of the process.
// A failed load denies every chat and is retried on the next message.
type Authorizer struct {
	loader  AllowedChatIdsLoader
	allowed map[telegram.ChatId]struct{}
	mu      sync.Mutex
}

func NewAuthorizer(loader AllowedChatIdsLoader) *Authorizer {
	return &Authorizer{loader: loader}
}

func (a *Authorizer) load() map[telegram.ChatId]struct{} {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.allowed != nil {
		return a.allowed
	}
	if a.loader == nil {
		log.Error("No allowed chat ids loader configured, every chat will be rejected")
		return nil
	}
	chatIds, err := a.loader()
	if err != nil {
		log.Errorf("Could not load allowed chat ids, every chat will be rejected: %v", err)
		return nil
	}
	allowed := make(map[telegram.ChatId]struct{}, len(chatIds))
	for _, chatId := range chatIds {
		allowed[telegram.ChatId(chatId)] = struct{}{}
	}
	if len(allowed) == 0 {
		log.Warn("Allowed chat ids list is empty, every chat will be rejected")
	}
	log.Infof("Loaded %d allowed chat ids", len(allowed))
	a.allowed = allowed
	return a.allowed
}

func (a *Authorizer) IsAllowed(chatId telegram.ChatId) bool {
	if a == nil {
		return false
	}
	_, ok := a.load()[chatId]
	return ok
}

func NewUnauthorizedProcessor(client *telegram.Client) UnauthorizedProcessor {
	return UnauthorizedProcessor{
		ProcessorShared: ProcessorShared{
			Client: client,
		},
	}
}

func unauthorizedMessage(chatId telegram.ChatId) string {
	return fmt.Sprintf("⛔ You are not authorized to use this bot. Your chat id is `%d`.", chatId)
}

// UnauthorizedProcessor replies to chats that are not in the allowlist
type UnauthorizedProcessor struct {
	ProcessorShared
}

func (p UnauthorizedProcessor) Process(args []string, update *telegram.Update) error {
	chatId := update.Message.Chat.ChatId
	log.Warnf("Rejecting message from unauthorized chat %d", chatId)
	err := p.Client.SendMessage(unauthorizedMessage(chatId), update.Message.Chat)
	if err != nil {
		log.Error("UnauthorizedProcessor: Error while sending message to Telegram")
		return err
	}
	return nil
}
