package processor

import (
	"fmt"

	"github.com/donkeysharp/donkeyvpn/internal/telegram"
	"github.com/labstack/gommon/log"
)

func NewCreateProcessor(client *telegram.Client) CreateProcessor {
	return CreateProcessor{
		ProcessorShared: ProcessorShared{
			Client: client,
		},
	}
}

type CreateProcessor struct {
	ProcessorShared
}

func (p CreateProcessor) Process(args []string, update *telegram.Update) {
	log.Infof("Processing '/create' command with args %v for chat %d", args, update.Message.Chat.ChatId)
	p.Client.SendMessage(
		fmt.Sprintf("Processed '/create' command with args %v", args),
		update.Message.Chat,
	)
}
