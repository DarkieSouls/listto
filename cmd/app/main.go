package main

import (
	"fmt"

	"github.com/DarkieSouls/listto/cmd/config"
	"github.com/DarkieSouls/listto/internal/bot"
)

func main() {
	config, err := config.NewConfig()
	if err != nil {
		panic(fmt.Sprintf("%s: %s", err.CallingMethod(), err.Error()))
	}

	bot.Start(config)

	<-make(chan struct{})
}
