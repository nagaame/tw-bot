package main

import (
	"fmt"
	"github.com/takama/daemon"
	"tw-bot/config"
	"tw-bot/database"
	"tw-bot/telegram"
	"tw-bot/twitter"
)

func main() {

	service, err := daemon.New("tw-bot", "A Telegram bot for Twitter", daemon.UserAgent)
	if err != nil {
		panic(err)
	}
	status, err := service.Install()
	if err != nil {
		panic(err)

	}
	fmt.Println(status)
	config.StartConfig()
	database.LoadData()
	twitter.Start()
	telegram.Start()
	select {}
}
