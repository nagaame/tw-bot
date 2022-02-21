package telegram

import (
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"sync"
	"time"
	"tw-bot/entity"
	"tw-bot/pkg"
)

func Send() {
	bot, err := tgApi.NewBotAPI("5084700957:AAEYjwCOopM7N0tmD63TOyVDrm8gLlsUMxY")
	if err != nil {
		log.Fatalln(err)
	}
	bot.Debug = true
	var idCh chan string
	err = pkg.Subscribe("twitter", func(message string) {
		if message == "" {
			return
		}
		idCh <- message
	})
	if err != nil {
		log.Fatalln(err)
	}
	for {
		select {
		case message := <-idCh:
			SendMessage(message, bot)
			time.Sleep(time.Second * 5)
		}
	}
}

func SendMessage(idStr string, bot *tgApi.BotAPI) {
	var err error
	db := pkg.GetDB()
	row := db.QueryRow("select * from tweets where tid = ?", idStr)
	if err != nil {
		return
	}
	item := entity.BotTweet{}
	err = row.Scan(&item.ID, &item.Content, &item.Tags, &item.MediaUrls, &item.Author)
	wg := sync.WaitGroup{}

	wg.Add(1)

	go func() {
		defer wg.Done()
		time.Sleep(time.Second * 10)
		_, err := bot.Send(tgApi.NewMessage(-1001278086217, item.Content))
		if err != nil {
			return
		}
	}()

	wg.Wait()

}
