package processor

import (
	"fmt"

	"github.com/donkeysharp/donkeyvpn/internal/telegram"
	"github.com/labstack/gommon/log"
)

func NewListProcessor(client *telegram.Client) ListProcessor {
	return ListProcessor{
		ProcessorShared: ProcessorShared{
			Client: client,
		},
	}
}

type ListProcessor struct {
	ProcessorShared
}

func (p ListProcessor) Process(args []string, update *telegram.Update) {
	log.Infof("Processing '/list' command with args %v for chat %d", args, update.Message.Chat.ChatId)
	p.Client.SendMessage(
		fmt.Sprintf("Processed '/list' command with args %v", args),
		update.Message.Chat,
	)
}
