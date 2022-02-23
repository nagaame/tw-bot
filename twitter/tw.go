package twitter

import (
	"context"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/duke-git/lancet/convertor"
	"strconv"
	"tw-bot/cache"
	"tw-bot/database"
	"tw-bot/entity"
)

type Twitter struct {
	client *twitter.Client
	tweets []entity.Tweets
}

//var (
//	tweets = make([]entity.Tweets, 0)
//)

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

func (t *Twitter) Fetch(ch chan bool) {
	client := NewTwitterClient()
	tweets, resp, err := client.Favorites.List(&twitter.FavoriteListParams{
		Count: 20,
	})
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != 200 {
		panic(resp.Status)
	}
	t.Collation(tweets)
}

func (t *Twitter) Collation(tweets []twitter.Tweet) {

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

func (t *Twitter) SaveToDB() {

	db := database.GetDataBase()

	for _, t := range t.tweets {
		if db.IsExists(t.ID) {
			return
		}
		id, err := db.SaveToDB(t.ID, t.Author, t.Content, t.Tags, t.MediaUrls, t.Url)
		if err != nil {
			return
		}
		if id == 0 {
			return
		}
	}
}

func (t *Twitter) SaveToCache() {
	c := cache.NewCache()
	for _, t := range t.tweets {
		idStr := strconv.FormatInt(t.ID, 10)
		_, err := c.SAdd("entity", idStr)
		if err != nil {
			return
		}
	}

}
