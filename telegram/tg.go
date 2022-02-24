package telegram

import (
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"tw-bot/cache"
	"tw-bot/database"
	"tw-bot/tool"
)

type TGBot struct {
	bot *tgApi.BotAPI
}

func NewTGBot(token string) *TGBot {
	bot, err := tgApi.NewBotAPI(token)
	bot.Debug = true
	if err != nil {
		log.Println(err)
	}
	return &TGBot{bot: bot}
}

func Start() {
	t := GetTGBot()
	t.Publish()
	t.Send()
}

func GetTGBot() *TGBot {
	return NewTGBot("5084700957:AAEYjwCOopM7N0tmD63TOyVDrm8gLlsUMxY")
}
func (t *TGBot) Send() {
	// send message
	redis := cache.NewRedisCache()
	pubSub := redis.Subscribe("twitter")
	go redis.HandlerSubscribe(pubSub, Handler)
}

func Handler(channel, payload string) {
	// handle message
	log.Println("channel: ", channel, "payload: ", payload)
	// send message
	t := GetTGBot()
	t.SendMessage(payload)
}

func (t *TGBot) Publish() {
	//ticker := time.NewTicker(time.Second * 10)

}

func (t *TGBot) SendMessage(idStr string) {
	db := database.GetDataBase()
	id := tool.StringToInt(idStr)
	tweet, err := db.QueryOne(id)
	if err != nil {
		log.Println(err)
		return
	}
	_, err = t.bot.Send(tgApi.NewMessage(-1001278086217, tweet.Content))
	if err != nil {
		return
	}

}
