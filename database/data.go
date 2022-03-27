package database

import (
	"fmt"
	"tw-bot/cache"
	"tw-bot/keys"
	"tw-bot/tool"
)

func LoadData() {

	db := GetDataBase()
	rows, err := db.QueryAll()
	if err != nil {
		panic(err)
	}
	c := cache.NewRedisCache()

	var counter int

	for index, tweet := range rows {
		idStr := tool.IntToString(tweet.ID)
		_, err := c.SAdd(keys.MainCacheTweets, idStr)
		if err != nil {
			continue
		}
		counter = index + 1
	}
	fmt.Println("Total tweets: ", counter)

}
