package twitter

import (
	"context"
	"fmt"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/duke-git/lancet/convertor"
	"github.com/rs/xid"
	"log"
	"net/http"
	"sync"
	"time"
	"tw-bot/cache"
	"tw-bot/data"
	"tw-bot/database"
	"tw-bot/tool"
)

var (
	MqSaveToDataBase      = "mq_save_to_database"
	MqSaveToDataBaseGroup = "mq_save_to_database_group"
	TweetToMQ             = "tweet_to_mq"
	TweetToMQCustomer     = "tweet_to_mq_customer"
	PushMQ                []data.Tweets
)

type Twitter struct {
	client *twitter.Client
	tweets []data.Tweets
}

func NewTwitter() *Twitter {
	return &Twitter{
		tweets: make([]data.Tweets, 0),
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
	go AsyncFetch()
	go PushMessage()
	//go t.AsyncPublishDataBaseMessage()
	//go t.SaveToDataBase()
}

func AsyncFetch() {
	ticker := time.NewTicker(time.Second * 60)
	// 轮询 每分钟一次
	for {
		Fetch()
		<-ticker.C
	}
}
func Fetch() {
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
	t := NewTwitter()
	t.Convert(&list)
}

func (t *Twitter) Convert(twitterList *[]twitter.Tweet) {

	for _, value := range *twitterList {
		tw := data.Tweets{}
		tw.ID = value.ID
		tw.Author = value.User.Name
		if len(value.Entities.Media) > 0 {
			tw.Url = value.Entities.Media[0].URL
		} else {
			tw.Url = ""
		}
		tw.Content = value.Text
		tempUrls := make([]string, 0)
		tempTags := make([]string, 0)
		if len(value.ExtendedEntities.Media) > 0 {
			for _, media := range value.ExtendedEntities.Media {
				tempUrls = append(tempUrls, media.MediaURL)
			}
		}
		if len(value.Entities.Hashtags) > 0 {
			for _, hashtag := range value.Entities.Hashtags {
				tempTags = append(tempTags, hashtag.Text)
			}
		}
		tw.MediaUrls = convertor.ToString(tempUrls)
		tw.Tags = convertor.ToString(tempTags)
		t.tweets = append(t.tweets, tw)
	}
	t.SaveToRedis()
}

func PushMessage() {
	ticker := time.NewTicker(time.Second * 45)
	// 轮询 每分钟一次
	for {
		PushToMQ()
		<-ticker.C
	}
}

func (t *Twitter) SaveToRedis() {
	c := cache.NewRedisCache()
	members, err := c.SMembers("tweets")
	if err != nil {
		fmt.Println(err)
	}
	if len(members) <= 0 {
		for _, value := range t.tweets {
			idStr := tool.IntToString(value.ID)
			_, err = c.SAdd("tweets", idStr)
			if err != nil {
				fmt.Println(err)
			}
		}

	}
	t.Difference()
}

func (t *Twitter) Difference() {
	c := cache.NewRedisCache()
	newKeys := make([]string, 0)
	for _, value := range t.tweets {
		idStr := tool.IntToString(value.ID)
		newKeys = append(newKeys, idStr)
	}
	_, err := c.SAdd("temp_tweets", newKeys)
	if err != nil {
		fmt.Println(err)
	}
	// 查询两个集合的差集
	diff, err := c.SDiff("tweets", "temp_tweets")
	if err != nil {
		fmt.Println(err)
	}
	if len(diff) == 0 {
		_ = c.Del("temp_tweets")
		if len(PushMQ) == 0 {
			PushMQ = t.tweets
		}
		return
	}
	diffTweets := make([]data.Tweets, 0)
	for _, value := range diff {
		tweets := t.GetTweet(tool.StringToInt(value))
		diffTweets = append(diffTweets, *tweets)
	}

	if err != nil {
		log.Println(err.Error())
	}
	PushMQ = diffTweets
}

func PushToMQ() {
	c := cache.NewRedisCache()
	once := sync.Once{}
	once.Do(func() {
		group, _ := c.XGroupCreate(TweetToMQ, TweetToMQCustomer)
		fmt.Println("group created is :", group)
	})
	for _, value := range PushMQ {
		//发送消息到消息队列
		id, err := c.XAdd(TweetToMQ, value)
		if err != nil {
			return
		}
		fmt.Println(id)
	}
	PushMQ = []data.Tweets{}

}

func (t *Twitter) GetTweets() []data.Tweets {
	return t.tweets
}

func (t *Twitter) GetTweet(id int64) *data.Tweets {
	for index, value := range t.tweets {
		if value.ID == id {
			return &t.tweets[index]
		}
	}
	return nil
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
			for _, message := range item.Messages {
				message := message.Values["tid"]
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
