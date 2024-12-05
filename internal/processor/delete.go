package processor

import (
	"fmt"

	"github.com/donkeysharp/donkeyvpn/internal/telegram"
	"github.com/labstack/gommon/log"
)

func NewDeleteProcessor(client *telegram.Client) DeleteProcessor {
	return DeleteProcessor{
		ProcessorShared: ProcessorShared{
			Client: client,
		},
	}
}

type DeleteProcessor struct {
	ProcessorShared
}

func (p DeleteProcessor) Process(args []string, update *telegram.Update) error {
	log.Infof("Processing '/delete' command with args %v for chat %d", args, update.Message.Chat.ChatId)
	// TODO check err
	p.Client.SendMessage(
		fmt.Sprintf("Processed '/delete' command with args %v", args),
		update.Message.Chat,
	)
	return nil
}
