package main

import (
	"tw-bot/telegram"
	"tw-bot/twitter"
)

func main() {

	ch := make(chan bool)
	t := twitter.NewTwitter()
	t.Fetch(ch)

	telegram.Publish(ch)
	telegram.Send()

}
