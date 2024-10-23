package processor

import "github.com/donkeysharp/donkeyvpn/internal/telegram"

type ProcessorShared struct {
	Client *telegram.Client
}

type Processor interface {
	Process(args []string, update *telegram.Update)
}
