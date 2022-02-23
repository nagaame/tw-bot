package telegram

import (
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"sync"
	"time"
	"tw-bot/cache"
	"tw-bot/database"
)

var (
	bot *tgApi.BotAPI
)

func init() {
	var err error
	bot, err = tgApi.NewBotAPI("5084700957:AAEYjwCOopM7N0tmD63TOyVDrm8gLlsUMxY")
	if err != nil {
		log.Fatalln(err)
	}
	bot.Debug = true
}

func Send() {
	// send message
	cache := cache.NewCache()

	pubSub := cache.Subscribe("twitter")

	go cache.HandlerSubscribe(pubSub)

}

func Publish(ch chan bool) {
	for {
		ok := <-ch
		if ok {
			cache := cache.NewCache()
			id, err := cache.SPop("entity")
			if err != nil {
				log.Println("redis key is empty: ", err.Error())
				continue
			}
			err = cache.Publish("twitter", id)
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}
}

func SendMessage(idStr string, bot *tgApi.BotAPI) {
	db := database.GetDataBase()
	id, _ := strconv.ParseInt(idStr, 10, 64)
	item, err := db.QueryOne(id)
	if err != nil {
		log.Println(err)
		return
	}
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
