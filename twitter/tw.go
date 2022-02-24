package twitter

import (
	"context"
	"fmt"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/duke-git/lancet/convertor"
	"github.com/go-redis/redis/v8"
	"github.com/rs/xid"
	"log"
	"net/http"
	"sync"
	"time"
	"tw-bot/cache"
	"tw-bot/database"
	"tw-bot/entity"
	"tw-bot/tool"
)

var (
	RedisCacheTweetsKey   = "tweets"
	MqSaveToDataBase      = "mq_save_to_database"
	MqSaveToDataBaseGroup = "mq_save_to_database_group"
	TweetToMQ             = "tweet_to_mq"
	TweetToMQCustomer     = "tweet_to_mq_customer"
)

type Twitter struct {
	client *twitter.Client
	tweets []entity.Tweets
}

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
	go t.AsyncFetch()
	//go t.AsyncPublishDataBaseMessage()
	//go t.SaveToDataBase()
}

func (t *Twitter) AsyncFetch() {
	ticker := time.NewTicker(time.Second * 60)
	// 轮询
	for {
		t.Fetch()
		<-ticker.C
	}
}
func (t *Twitter) Fetch() {
	client := NewTwitterClient()
	list := make([]twitter.Tweet, 0)
	var err error
	var resp *http.Response
	list, resp, err = client.Favorites.List(&twitter.FavoriteListParams{
		Count: 20,
	})
	if err != nil {
		log.Println(err)
		return
	}
	if resp.StatusCode != 200 {
		log.Println(err)
		return
	}
	t.Convert(&list)
}

func (t *Twitter) Convert(twitterTweets *[]twitter.Tweet) {
	for _, value := range *twitterTweets {
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
		t.tweets = append(t.tweets, bt)
	}

	t.SaveToRedis()
	for _, value := range t.tweets {
		PushToMQ(value)
	}
	// 销毁
	twitterTweets = nil
	fmt.Println("twitter fetch once success")
}

func PushToMQ(t entity.Tweets) {
	c := cache.NewRedisCache()
	once := sync.Once{}
	once.Do(func() {
		group, _ := c.XGroupCreate(TweetToMQ, TweetToMQCustomer)
		fmt.Println(group)
	})
	idStr := tool.IntToString(t.ID)
	exist, err := c.SIsMember(RedisCacheTweetsKey, idStr)
	if err != nil {
		return
	}
	if exist {
		return
	}

	id, err := c.XAdd(TweetToMQ, t)
	if err != nil {
		return
	}
	fmt.Println(id)
}

func (t *Twitter) AsyncPublishDataBaseMessage() {
	c := cache.NewRedisCache()
	ticker := time.NewTicker(time.Second * 3)
	for {
		idStr, err := c.SPop(RedisCacheTweetsKey)
		if err == redis.Nil {
			continue
		}
		messageId, err := c.XAdd(MqSaveToDataBase, idStr)
		if err != nil {
			continue
		}
		fmt.Println(messageId)
		<-ticker.C
	}

}

func (t *Twitter) GetTweets() []entity.Tweets {
	return t.tweets
}

func (t *Twitter) GetTweet(id int64) *entity.Tweets {
	for index, value := range t.tweets {
		if value.ID == id {
			return &t.tweets[index]
		}
	}
	return nil
}

func (t *Twitter) SaveToRedis() {
	c := cache.NewRedisCache()
	for _, item := range t.tweets {
		idStr := tool.IntToString(item.ID)
		_, err := c.SAdd(RedisCacheTweetsKey, idStr)
		if err != nil {
			return
		}
	}
}

func (t *Twitter) SaveToDataBase() {
	c := cache.NewRedisCache()
	ticker := time.NewTicker(time.Second * 3)
	//group, err := c.XGroupCreate(MqSaveToDataBase, MqSaveToDataBaseGroup)
	//if group != "OK" || err != nil {
	//	return
	//}
	for {

		uniqueID := xid.New().String()
		result, err := c.XReadGroup(MqSaveToDataBase, MqSaveToDataBaseGroup, uniqueID, 1)
		if err != nil {
			return
		}
		for _, item := range result {
			for _, data := range item.Messages {
				message := data.Values["tid"]
				id := tool.StringToInt(message.(string))
				tweet := t.GetTweet(id)
				db := database.GetDataBase()
				_, err = db.SaveOne(*tweet)
				if err != nil {
					return
				}
			}
		}
		<-ticker.C
	}

}
