package processor

import (
	"fmt"

	"github.com/donkeysharp/donkeyvpn/internal/telegram"
	"github.com/labstack/gommon/log"
)

func NewDocsProcessor(client *telegram.Client) *DocsProcessor {
	return &DocsProcessor{
		ProcessorShared: ProcessorShared{
			Client: client,
		},
	}
}

type DocsProcessor struct {
	ProcessorShared
}

func (p *DocsProcessor) Process(args []string, update *telegram.Update) error {
	log.Info("Processing '/docs' command")

	log.Infof("Generate docs response")
	message := ""
	manualsUrl := "https://github.com/donkeysharp/donkeyvpn/tree/master?tab=readme-ov-file#manuals"
	message += fmt.Sprintf("*Manuals Url*: `%v\n`", manualsUrl)
	message += "-----\n"

	p.SendMessage(message, update)

	return nil
}
