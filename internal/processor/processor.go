package processor

import (
	"github.com/donkeysharp/donkeyvpn/internal/telegram"
	"github.com/labstack/gommon/log"
)

type ProcessorShared struct {
	Client *telegram.Client
}

type Processor interface {
	Process(args []string, update *telegram.Update) error
}

func (p ProcessorShared) SendMessage(msg string, update *telegram.Update) {
	err2 := p.Client.SendMessage(msg, update.Message.Chat)
	if err2 != nil {
		log.Errorf("Error sending message to Telegram. msg=%s", msg)
	}
}
