package twitter

import (
	"context"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"strconv"
	"strings"
	"tw-bot/entity"
	"tw-bot/pkg"
)

type Twitter struct {
	client *twitter.Client
}

var (
	botTweets = make([]entity.BotTweet, 0)
)

func Client() *twitter.Client {
	localConfig := GetConfig()
	config := oauth1.NewConfig(localConfig.ConsumerKey, localConfig.ConsumerSecret)
	token := oauth1.NewToken(localConfig.AccessToken, localConfig.AccessTokenSecret)
	httpClient := config.Client(context.Background(), token)
	client := twitter.NewClient(httpClient)
	return client
}

func Fetch() {
	client := Client()
	tweets, resp, err := client.Favorites.List(&twitter.FavoriteListParams{
		Count: 20,
	})
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != 200 {
		panic(resp.Status)
	}
	Collation(tweets)

	for _, b := range botTweets {
		SaveToCache(&b)
	}

	for _, b := range botTweets {
		SaveToDB(&b)
	}

	for _, b := range botTweets {
		idStr := strconv.FormatInt(b.ID, 10)
		err := pkg.Publish("twitter", idStr)
		if err != nil {
			return
		}
	}
}

func Collation(tweets []twitter.Tweet) []entity.BotTweet {
	for _, value := range tweets {
		bt := entity.BotTweet{}
		bt.ID = value.ID
		bt.Author = value.User.Name
		url := strings.Split(value.Text, " ")
		if len(url) > 1 {
			bt.Url = url[1]
		} else {
			bt.Url = ""
		}
		bt.Content = url[0]
		tempUrls := make([]string, 0)
		tempTags := make([]string, 0)
		for _, media := range value.ExtendedEntities.Media {
			tempUrls = append(tempUrls, media.MediaURLHttps)
		}
		for _, tag := range value.Entities.Hashtags {
			tempTags = append(tempTags, tag.Text)
		}
		bt.MediaUrls = strings.Join(tempUrls, ",")
		bt.Tags = strings.Join(tempTags, ",")

		botTweets = append(botTweets, bt)
	}

	return botTweets
}

func GetBotTweets() []entity.BotTweet {
	return botTweets
}

func GetBotTweet(id int64) entity.BotTweet {
	for _, value := range botTweets {
		if value.ID == id {
			return value
		}
	}
	return entity.BotTweet{}
}

func SaveToDB(bt *entity.BotTweet) {
	if pkg.IsExists(bt.ID) {
		return
	}
	id, err := pkg.SaveToDB(bt.ID, bt.Content, bt.Author, bt.MediaUrls, bt.Tags, bt.Url)
	if err != nil {
		return
	}
	if id == 0 {
		return
	}
}

func SaveToCache(bt *entity.BotTweet) {
	idStr := strconv.FormatInt(bt.ID, 10)
	ok, err := pkg.Exists(idStr)

	if err != nil {
		return
	}

	if ok > 0 {
		return
	}

	idStr = strconv.FormatInt(bt.ID, 10)
	err = pkg.Set(idStr, idStr)
	if err != nil {
		return
	}

}
