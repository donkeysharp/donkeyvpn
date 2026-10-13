package processor

import (
	"strings"

	"github.com/donkeysharp/donkeyvpn/internal/telegram"
	"github.com/labstack/gommon/log"
)

type Command string

const (
	UnknowCommand = "unknown"
)

func NewCommandProcessor(authorizer *Authorizer, unauthorized Processor) *CommandProcessor {
	return &CommandProcessor{
		processors:   make(map[Command]Processor),
		fallback:     UnknowCommandProcessor{},
		authorizer:   authorizer,
		unauthorized: unauthorized,
	}
}

type CommandProcessor struct {
	processors   map[Command]Processor
	fallback     Processor
	authorizer   *Authorizer
	unauthorized Processor
}

func (p *CommandProcessor) Register(command Command, processor Processor) {
	p.processors[command] = processor
}

func (p *CommandProcessor) RegisterFallback(processor Processor) {
	p.fallback = processor
}

func (p *CommandProcessor) Process(update *telegram.Update) {
	if !p.authorizer.IsAllowed(update.Message.Chat.ChatId) {
		err := p.unauthorized.Process(nil, update)
		if err != nil {
			log.Warnf("Error: %v", err)
		}
		return
	}

	command, args := parseCommand(update)
	log.Infof("Processing command: %s with args: %v", command, args)
	if processor, ok := p.processors[command]; ok {
		err := processor.Process(args, update)
		if err != nil {
			log.Warnf("%v failed to be processed, it will not interrupt execution", command)
			log.Warnf("Error: %v", err)
		}
		return
	}
	p.fallback.Process(args, update)
}

func parseCommand(update *telegram.Update) (Command, []string) {
	fields := strings.Fields(update.Message.Text)
	if len(fields) == 0 {
		return Command(UnknowCommand), nil
	}

	command := Command(fields[0])
	if len(fields) == 1 {
		return command, nil
	}
	return command, fields[1:]
}
