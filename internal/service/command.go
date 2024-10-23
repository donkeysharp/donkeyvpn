package service

import (
	"strings"

	"github.com/donkeysharp/donkeyvpn/internal/processor"
	"github.com/donkeysharp/donkeyvpn/internal/telegram"
	"github.com/labstack/gommon/log"
)

type Command string

const (
	UnknowCommand = "unknown"
)

func NewCommandService() *CommandService {
	return &CommandService{
		processors: make(map[Command]processor.Processor),
		fallback:   processor.UnknowCommandProcessor{},
	}
}

type CommandService struct {
	processors map[Command]processor.Processor
	fallback   processor.Processor
}

func (p *CommandService) Register(command Command, processor processor.Processor) {
	p.processors[command] = processor
}

func (p *CommandService) RegisterFallback(processor processor.Processor) {
	p.fallback = processor
}

func (p *CommandService) Process(update *telegram.Update) {
	command, args := parseCommand(update)
	log.Infof("Processing command: %s with args: %v", command, args)
	if processor, ok := p.processors[command]; ok {
		processor.Process(args, update)
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
