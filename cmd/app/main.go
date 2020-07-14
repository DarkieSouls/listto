package main

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"

	"github.com/DarkieSouls/listto/cmd/config"
	"github.com/DarkieSouls/listto/internal/bot"
)

func main() {
	config, err := config.NewConfig()
	if err != nil {
		err.LogError()
		os.Exit(-1)
	}

	awsCfg := aws.NewConfig().WithRegion("eu-west-2")
	sess := session.Must(session.NewSession(awsCfg))

	ddb := dynamodb.New(sess)

	bot := bot.New(config, ddb)

	bot.Start()

	<-make(chan struct{})
}
