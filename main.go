package main

import (
	"fmt"
	"time"
	"tw-bot/twitter"
)

func main() {
	twitter.Start()
	//telegram.Start()
	select {}

}

func test() {
	t := time.NewTicker(time.Second * 1)
	for {
		fmt.Println("test")
		<-t.C
	}

}
