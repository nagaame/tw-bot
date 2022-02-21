package main

import (
	"tw-bot/telegram"
	"tw-bot/twitter"
)

func main() {
	twitter.Fetch()
	telegram.Send()
}
