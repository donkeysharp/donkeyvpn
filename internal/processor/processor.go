package processor

import (
	"github.com/donkeysharp/donkeyvpn/internal/aws"
	"github.com/donkeysharp/donkeyvpn/internal/telegram"
)

type ProcessorShared struct {
	Client *telegram.Client
	asg    *aws.AutoscalingGroup
	table  *aws.DynamoDB
}

type Processor interface {
	Process(args []string, update *telegram.Update) error
}
