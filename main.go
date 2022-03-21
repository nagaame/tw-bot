package main

import (
	"tw-bot/config"
	"tw-bot/telegram"
	"tw-bot/twitter"
)

func main() {

	config.StartConfig()
	twitter.Start()
	telegram.Start()

	select {}

}
