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
	"tw-bot/tool"
)

var (
	TweetToMQ         = "tweet_to_mq"
	TweetToMQCustomer = "tweet_to_mq_customer"
	tg                *TGBot
	tweet             *data.Tweet
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
	bot.Debug = true
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
		HandlerSendMessage()

	}
}

func SendMessage() {
	c := cache.NewRedisCache()
	var read []redis.XStream
	exists, err := c.Exists(TweetToMQ)
	if exists != 1 {
		return
	}
	id := xid.New().String()

	_, _ = c.XGroupCreate(TweetToMQ, TweetToMQCustomer)

	read, err = c.XReadGroup(TweetToMQ, TweetToMQCustomer, id, 1)
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
	tweetS := data.Tweet{}
	err = json.Unmarshal([]byte(value), &tweetS)
	if err != nil {
		log.Println("json unmarshal error: ", err)
		return
	}
	tweet = &tweetS
	return
}

func HandlerSendMessage() {
	t := GetTGBot()

	if tweet == nil {
		log.Println("tweet is nil")
		return
	}

	mediaMsg := MediaMessage(*tweet)
	_, err := t.bot.SendMediaGroup(mediaMsg.(tgApi.MediaGroupConfig))

	if err != nil {
		log.Println("send message: ", err)
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
