package main

import (
	"tw-bot/config"
	"tw-bot/database"
	"tw-bot/telegram"
	"tw-bot/twitter"
)

func main() {
	config.StartConfig()
	database.LoadData()
	twitter.Start()
	telegram.Start()
	select {}
}
