package twitter

import (
	"context"
	"fmt"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/duke-git/lancet/convertor"
	"log"
	"net/http"
	"net/url"
	"time"
	"tw-bot/cache"
	"tw-bot/config"
	"tw-bot/data"
	"tw-bot/database"
	"tw-bot/keys"
	"tw-bot/tool"
)

var (
	WaitTweetsMQ []data.Tweet
)

type Twitter struct {
	client *twitter.Client
	tweets []data.Tweet
}

func NewTwitter() *Twitter {
	return &Twitter{
		tweets: make([]data.Tweet, 0),
		client: NewTwitterClient(),
	}
}
func NewTwitterClient() *twitter.Client {
	localConfig := config.GetConfig()
	customerKey := localConfig.Twitter.CustomerKey
	customerSecret := localConfig.Twitter.CustomerSecret
	accessToken := localConfig.Twitter.AccessToken
	accessSecret := localConfig.Twitter.AccessTokenSecret

	c := oauth1.NewConfig(customerKey, customerSecret)
	token := oauth1.NewToken(accessToken, accessSecret)
	httpClient := c.Client(context.Background(), token)
	client := twitter.NewClient(httpClient)
	return client
}

func Start() {
	go AsyncFetch()
	go PushMessage()
	//go AsyncLikedFetch()
}

func AsyncFetch() {
	ticker := tool.NewTicker(time.Second * 60)
	ticker.Start(func() {
		Fetch()
	})
}

func AsyncLikedFetch() {
	t := tool.NewTicker(time.Second * 5)
	t.Start(func() {
		FetchLiked()
	})

}

func FetchLiked() {
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "FetchLiked")
}

func Fetch() {

	t := NewTwitter()
	list := make([]twitter.Tweet, 0)
	var err error
	var resp *http.Response
	list, resp, err = t.client.Favorites.List(&twitter.FavoriteListParams{
		Count: 200,
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

func (t *Twitter) Convert(twitterList *[]twitter.Tweet) {

	for _, list := range *twitterList {
		tweet := data.Tweet{}
		tweet.ID = list.ID
		tweet.Author = list.User.Name
		if len(list.Entities.Media) > 0 {
			tweet.Url = list.Entities.Media[0].URL
		} else {
			tweet.Url = ""
		}
		tweet.Content = list.Text
		mediaUrls := make([]string, 0)
		tags := make([]string, 0)
		if list.ExtendedEntities == nil {
			continue
			//mediaUrls = append(mediaUrls, "")
		} else {
			for _, media := range list.ExtendedEntities.Media {
				mediaUrl, err := url.Parse(media.MediaURLHttps)
				if err != nil {
					log.Println("url parse error:", err)
					continue
				}
				params := url.Values{}
				params.Add("name", "large")
				//mediaUrl.Scheme = "https"
				mediaUrl.RawQuery = params.Encode()
				mediaUrls = append(mediaUrls, mediaUrl.String())
			}
		}
		if list.Entities.Hashtags == nil {
			tags = append(tags, "")
		} else {
			for _, tag := range list.Entities.Hashtags {
				tags = append(tags, tag.Text)
			}
		}
		tweet.MediaUrls = convertor.ToString(mediaUrls)
		tweet.Tags = convertor.ToString(tags)
		t.tweets = append(t.tweets, tweet)
	}
	t.SaveToRedis()
}

func PushMessage() {
	c := cache.NewRedisCache()
	group, _ := c.XGroupCreate(keys.TweetToMQ, keys.TweetToMQCustomer)
	fmt.Println("group created is :", group)

	// 轮询 每45秒一次
	ticker := tool.NewTicker(time.Second * 45)
	ticker.Start(func() {
		PushToMQ()
	})

}

func (t *Twitter) SaveToRedis() {
	c := cache.NewRedisCache()
	// 先备份旧的键值
	oldKeys, _ := c.SMembers(keys.MainCacheTweets)
	for _, value := range t.tweets {
		idStr := tool.IntToString(value.ID)
		// 加入新的键值
		_, err := c.SAdd(keys.MainCacheTweets, idStr)
		if err != nil {
			fmt.Println(err)
		}
	}
	t.Difference(oldKeys)
}

func (t *Twitter) Difference(oldKeys []string) {
	c := cache.NewRedisCache()
	if len(oldKeys) != 0 {
		// 写入到新的key中
		_, err := c.SAdd(keys.OldCacheTweets, oldKeys)
		if err != nil {
			fmt.Println(err)
		}
	}
	// 比较差集
	diff, err := c.SDiff(keys.MainCacheTweets, keys.OldCacheTweets)
	if err != nil {
		fmt.Println(err)
	}
	if len(diff) == 0 {
		_ = c.Del(keys.OldCacheTweets)
		return
	}
	diffTweets := make([]data.Tweet, 0)
	for _, value := range diff {
		tweets := t.GetTweet(tool.StringToInt(value))
		diffTweets = append(diffTweets, *tweets)
	}

	if err != nil {
		log.Println(err.Error())
	}

	saveToDB(diffTweets)
	WaitTweetsMQ = diffTweets
}

func PushToMQ() {
	c := cache.NewRedisCache()

	for _, value := range WaitTweetsMQ {
		//发送消息到消息队列
		id, err := c.XAdd(keys.TweetToMQ, value)
		if err != nil {
			return
		}
		fmt.Println(id)
	}
	WaitTweetsMQ = []data.Tweet{}

}

func (t *Twitter) GetTweets() []data.Tweet {
	return t.tweets
}

func (t *Twitter) GetTweet(id int64) *data.Tweet {
	for index, value := range t.tweets {
		if value.ID == id {
			return &t.tweets[index]
		}
	}
	return nil
}

func saveToDB(tweets []data.Tweet) {
	db := database.GetDataBase()
	for _, tweet := range tweets {
		_, err := db.SaveOne(tweet)
		if err != nil {
			continue
		}
	}
}
