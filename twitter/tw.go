package twitter

import (
	"context"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/duke-git/lancet/convertor"
	"log"
	"strconv"
	"time"
	"tw-bot/cache"
	"tw-bot/database"
	"tw-bot/entity"
)

type Twitter struct {
	client *twitter.Client
	tweets []entity.Tweets
}

var (
	saveToCache chan bool
	saveToDB    chan bool
)

func NewTwitter() *Twitter {
	return &Twitter{
		tweets: make([]entity.Tweets, 0),
		client: NewTwitterClient(),
	}
}

func NewTwitterClient() *twitter.Client {
	localConfig := GetConfig()
	config := oauth1.NewConfig(localConfig.ConsumerKey, localConfig.ConsumerSecret)
	token := oauth1.NewToken(localConfig.AccessToken, localConfig.AccessTokenSecret)
	httpClient := config.Client(context.Background(), token)
	client := twitter.NewClient(httpClient)
	return client
}

func Start() {
	t := NewTwitter()
	t.TickerFetch()

	for {
		select {
		case <-saveToCache:
			t.SaveToRedis()
		case <-saveToDB:
			t.SaveToDataBase()
		}
	}
}

func (t *Twitter) TickerFetch() {
	ticker := time.NewTicker(time.Second * 5)
	redis := cache.NewCache()
	maxId, err := redis.Get("max_id")
	if err != nil {
		log.Println(err)
	}
	if maxId == "" {
		maxId = "0"
	}
	id, _ := strconv.ParseInt(maxId, 10, 64)

	for {
		t.Fetch(id)
		<-ticker.C
	}
}
func (t *Twitter) Fetch(maxId int64) {
	client := NewTwitterClient()
	tweets, resp, err := client.Favorites.List(&twitter.FavoriteListParams{
		Count:   20,
		SinceID: maxId,
	})
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != 200 {
		panic(resp.Status)
	}
	t.Convert(tweets)
	maxId = t.MaxId()
	redis := cache.NewCache()
	_, err = redis.Set("max_id", strconv.FormatInt(maxId, 10))
	if err != nil {
		log.Println(err.Error())
		return
	}
}

func (t *Twitter) MaxId() int64 {
	var id int64
	for _, tweet := range t.tweets {
		if tweet.ID > id {
			id = tweet.ID
		}
	}
	return id
}

func (t *Twitter) MaxIdPublish(m int64) error {
	redis := cache.NewCache()
	err := redis.Publish("max_id", strconv.FormatInt(m, 10))
	if err != nil {
		return err
	}
	return nil
}

func (t *Twitter) Convert(tweets []twitter.Tweet) {

	ts := make([]entity.Tweets, 0)

	for _, value := range tweets {
		bt := entity.Tweets{}
		bt.ID = value.ID
		bt.Author = value.User.Name
		if len(value.Entities.Media) > 0 {
			bt.Url = value.Entities.Media[0].URL
		} else {
			bt.Url = ""
		}
		bt.Content = value.Text
		tempUrls := make([]string, 0)
		tempTags := make([]string, 0)
		for _, media := range value.ExtendedEntities.Media {
			tempUrls = append(tempUrls, media.MediaURLHttps)
		}
		for _, tag := range value.Entities.Hashtags {
			tempTags = append(tempTags, tag.Text)
		}
		bt.MediaUrls = convertor.ToString(tempUrls)
		bt.Tags = convertor.ToString(tempTags)
		ts = append(ts, bt)
	}

	t.tweets = ts
	saveToCache <- true
	saveToDB <- true
}

func (t *Twitter) GetTweets() []entity.Tweets {
	return t.tweets
}

func (t *Twitter) GetTweet(id int64) entity.Tweets {
	for _, value := range t.tweets {
		if value.ID == id {
			return value
		}
	}
	return entity.Tweets{}
}

func (t *Twitter) SaveToDataBase() {

	db := database.GetDataBase()

	for _, t := range t.tweets {
		if db.IsExists(t.ID) {
			return
		}
		id, err := db.SaveOne(t.ID, t.Author, t.Content, t.Tags, t.MediaUrls, t.Url)
		if err != nil {
			return
		}
		if id == 0 {
			return
		}
	}
}

func (t *Twitter) SaveToRedis() {
	c := cache.NewCache()
	for _, t := range t.tweets {
		idStr := strconv.FormatInt(t.ID, 10)
		_, err := c.SAdd("tweets", idStr)
		if err != nil {
			return
		}
	}
}
