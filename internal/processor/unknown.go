package processor

import (
	"fmt"

	"github.com/donkeysharp/donkeyvpn/internal/telegram"
	"github.com/labstack/gommon/log"
)

func NewUnknowCommandProcessor(client *telegram.Client) UnknowCommandProcessor {
	return UnknowCommandProcessor{
		ProcessorShared: ProcessorShared{
			Client: client,
		},
	}
}

type UnknowCommandProcessor struct {
	ProcessorShared
}

func (p UnknowCommandProcessor) Process(args []string, update *telegram.Update) error {
	log.Info("Processing unknown command")
	message := fmt.Sprintf("Sorry, invalid command\nUsage:\n%s", getUsage())
	err := p.Client.SendMessage(message, update.Message.Chat)
	if err != nil {
		log.Error("UnknowCommandProcessor: Error while sending message to Telegram")
		return err
	}
	return nil
}
