package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/donkeysharp/donkeyvpn/internal/app"
	"github.com/donkeysharp/donkeyvpn/internal/config"
	"github.com/donkeysharp/donkeyvpn/internal/telegram"
	"github.com/labstack/gommon/log"
)

type TaskType string

const TASK_CHECK_INSTANCES TaskType = "checkInstances"

type ScheduledTask struct {
	Task TaskType `json:"task"`
}

func handleRequest(ctx context.Context, event json.RawMessage) error {
	var task ScheduledTask
	if err := json.Unmarshal(event, &task); err != nil {
		log.Errorf("Error while parsing event")
		return err
	}

	log.Infof("Task: %v", task.Task)

	if task.Task == TASK_CHECK_INSTANCES {
		log.Infof("Processing task")

	} else {
		log.Infof("Unknown task")
	}

	cfg := config.LoadConfigFromEnvVars()
	client := telegram.NewClient(cfg.TelegramBotAPIToken)
	notifier, err := app.NewNotifierApplication(cfg, client)
	if err != nil {
		return err
	}

	if err := notifier.CheckInstances(); err != nil {
		log.Errorf("Failed to check instances: %v", err.Error())
		return err
	}

	return nil
}

func main() {
	var runAsLambdaStr string = os.Getenv("RUN_AS_LAMBDA")
	var runAsLambda bool = runAsLambdaStr == "true"
	if runAsLambda {
		log.Info("Running handler in Lambda function")
		lambda.Start(handleRequest)
	} else {
		log.Info("Running handler request locally")
		sample := "{\"task\": \"checkInstances\"}"
		handleRequest(context.Background(), []byte(sample))
	}
}
