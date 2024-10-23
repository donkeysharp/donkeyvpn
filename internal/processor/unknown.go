package processor

import (
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

func (p UnknowCommandProcessor) Process(args []string, update *telegram.Update) {
	log.Info("Processing unknown command")
	p.Client.SendMessage("Sorry, invalid command", update.Message.Chat)
}
