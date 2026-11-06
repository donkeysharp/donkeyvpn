package testutil

import (
	"fmt"
	"sync"
	"time"

	asgtypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/donkeysharp/donkeyvpn/internal/telegram"
)

// FakeScalingGroup has the same methods as *aws.AutoscalingGroup and keeps
// the desired capacity in memory.
type FakeScalingGroup struct {
	mu               sync.Mutex
	DesiredCapacity  int32
	MaxSize          int32
	CapacityUpdates  []int32
	DeletedInstances []string
	Err              error
}

func (f *FakeScalingGroup) GetInfo() (*asgtypes.AutoScalingGroup, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.Err != nil {
		return nil, f.Err
	}
	desired, max := f.DesiredCapacity, f.MaxSize
	return &asgtypes.AutoScalingGroup{DesiredCapacity: &desired, MaxSize: &max}, nil
}

func (f *FakeScalingGroup) UpdateCapacity(desiredCapacity int32) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.Err != nil {
		return f.Err
	}
	f.CapacityUpdates = append(f.CapacityUpdates, desiredCapacity)
	f.DesiredCapacity = desiredCapacity
	return nil
}

// DeleteInstance records the id and decrements the desired capacity, like the
// real call with ShouldDecrementDesiredCapacity.
func (f *FakeScalingGroup) DeleteInstance(instanceId string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.Err != nil {
		return f.Err
	}
	f.DeletedInstances = append(f.DeletedInstances, instanceId)
	if f.DesiredCapacity > 0 {
		f.DesiredCapacity--
	}
	return nil
}

// FakeInstanceDescriber has the same methods as *aws.EC2. Unknown ids return
// an error, like the real client.
type FakeInstanceDescriber struct {
	Instances map[string]ec2types.Instance
	Err       error
}

func (f *FakeInstanceDescriber) DescribeInstance(instanceId string) (*ec2types.Instance, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	instance, ok := f.Instances[instanceId]
	if !ok {
		return nil, fmt.Errorf("no instance found with the given id")
	}
	return &instance, nil
}

// NewEC2Instance builds an instance with the fields VPNService reads.
func NewEC2Instance(instanceId string, state ec2types.InstanceStateName, launchTime time.Time) ec2types.Instance {
	return ec2types.Instance{
		InstanceId: &instanceId,
		State:      &ec2types.InstanceState{Name: state},
		LaunchTime: &launchTime,
	}
}

// SentMessage is a message passed to RecordingSender.SendMessage, before any
// MarkdownV2 escaping.
type SentMessage struct {
	ChatId telegram.ChatId
	Text   string
}

// RecordingSender has the same SendMessage method as *telegram.Client and
// records each message instead of calling Telegram.
type RecordingSender struct {
	mu       sync.Mutex
	messages []SentMessage
	Err      error
}

func (s *RecordingSender) SendMessage(message string, chat *telegram.Chat) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, SentMessage{ChatId: chat.ChatId, Text: message})
	return s.Err
}

// Messages returns a copy of the recorded messages in the order they were sent.
func (s *RecordingSender) Messages() []SentMessage {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]SentMessage(nil), s.messages...)
}
