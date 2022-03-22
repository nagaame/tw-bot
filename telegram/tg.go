package telegram

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	tgApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/xid"
	"log"
	"strings"
	"time"
	"tw-bot/cache"
	"tw-bot/config"
	"tw-bot/data"
	"tw-bot/database"
	"tw-bot/tool"
	"tw-bot/twitter"
)

var (
	tg *TGBot
)

type TGBot struct {
	bot *tgApi.BotAPI
}

func NewTGBot(token string) *TGBot {
	if tg != nil {
		return tg
	}
	tg = new(TGBot)
	bot, err := tgApi.NewBotAPI(token)
	bot.Debug = false
	if err != nil {
		log.Println(err)
	}
	tg.bot = bot
	return tg
}

func Start() {
	go Subscribe()
}

func GetTGBot() *TGBot {
	token := config.GetConfig().Telegram.TGBotToken
	return NewTGBot(token)
}

func Subscribe() {
	// subscribe
	ticker := time.NewTicker(time.Second * 30)
	for {
		<-ticker.C
		SendMessage()
	}
}

func SendMessage() {
	c := cache.NewRedisCache()
	exists, err := c.Exists(twitter.TweetToMQ)
	if exists != 1 {
		return
	}
	id := xid.New().String()
	_, _ = c.XGroupCreate(twitter.TweetToMQ, twitter.TweetToMQCustomer)
	read, err := c.XReadGroup(twitter.TweetToMQ, twitter.TweetToMQCustomer, id, 1)
	if err != nil {
		log.Println("x read group error:", err)
		return
	}
	stream := read[0]
	message := stream.Messages[0]

	value, ok := message.Values["tweet"].(string)
	if !ok {
		log.Println("value is not string")
		return
	}

	tweet := data.Tweet{}
	err = json.Unmarshal([]byte(value), &tweet)
	if err != nil {
		log.Println("json unmarshal error: ", err)
		return
	}
	mediaMsg := MediaMessage(tweet)
	t := GetTGBot()
	_, err = t.bot.SendMediaGroup(mediaMsg.(tgApi.MediaGroupConfig))

	if err != nil {
		log.Println("send message: ", err)
		DeleteStreamMessage(stream)
		return
	}
	return
}

func DeleteStreamMessage(stream redis.XStream) {
	c := cache.NewRedisCache()
	message := stream.Messages[0]
	id := message.ID
	_, err := c.XAck(twitter.TweetToMQ, twitter.TweetToMQCustomer, id)
	if err != nil {
		log.Println("x ack error: ", err)
	}
	value, ok := message.Values["tweet"].(string)
	if !ok {
		log.Println("value is not string")
		return
	}
	tweet := data.Tweet{}
	err = json.Unmarshal([]byte(value), &tweet)
	if err != nil {
		log.Println("json unmarshal error: ", err)
		return
	}
	CleanCacheAndDB(tweet)
}

func CleanCacheAndDB(tweet data.Tweet) {
	c := cache.NewRedisCache()
	_, err := c.SRem(twitter.MainCacheTweets, tool.IntToString(tweet.ID))
	if err != nil {
		log.Println("s rem error: ", err)
		return
	}
	db := database.GetDataBase()
	err = db.Delete(tweet.ID)
	if err != nil {
		log.Println("delete error: ", err)
		return
	}
}

func CaptionFormat(tweet data.Tweet) string {
	formatStr := ""
	var tags []string
	_ = json.Unmarshal([]byte(tweet.Tags), &tags)
	if len(tags) > 0 {
		tagStr := ""

		for _, tag := range tags {
			tagStr += fmt.Sprintf("#%s ", tag)
		}
		tagStr = strings.TrimRight(tagStr, " ")
		formatStr = fmt.Sprintf("Title: %s\nTag: %s\nAuthor: %s", tweet.Content, tagStr, tweet.Author)
	} else {
		return fmt.Sprintf("Title: %s\nAuthor: %s", tweet.Content, tweet.Author)
	}
	return formatStr
}

func MediaMessage(t data.Tweet) tgApi.Chattable {
	var tagsGroup []string
	var mediasGroup []string
	err := json.Unmarshal([]byte(t.MediaUrls), &mediasGroup)
	if err != nil {
		log.Println(err)
	}
	err = json.Unmarshal([]byte(t.Tags), &tagsGroup)
	if err != nil {
		log.Println(err)
	}
	var medias []interface{}
	for index, mediaUrl := range mediasGroup {
		photo := tgApi.InputMediaPhoto{}
		if index == 0 {
			caption := CaptionFormat(t)
			photo = tgApi.NewInputMediaPhoto(tgApi.FileURL(mediaUrl))
			photo.Caption = caption
			medias = append(medias, photo)
		} else {
			photo = tgApi.NewInputMediaPhoto(tgApi.FileURL(mediaUrl))
			medias = append(medias, photo)
		}
	}
	channelID := config.GetConfig().Telegram.ChannelID
	id := tool.StringToInt(channelID)
	mediaMsg := tgApi.NewMediaGroup(id, medias)
	return mediaMsg
}
