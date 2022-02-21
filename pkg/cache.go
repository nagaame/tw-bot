package pkg

import (
	"context"
	"github.com/go-redis/redis"
	"strconv"
)

var cache *redis.Client

func init() {
	cache = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func Set(key string, value string) error {
	return cache.Set(key, value, 0).Err()
}

func Get(key string) (string, error) {
	return cache.Get(key).Result()
}

func Del(key string) error {
	return cache.Del(key).Err()
}

func Flush() error {
	return cache.FlushDB().Err()
}

func Exists(key string) (int64, error) {
	return cache.Exists(key).Result()
}

func Publish(channel string, message string) error {
	return cache.Publish(channel, message).Err()
}

func Subscribe(channel string, handler func(string)) error {
	sub := cache.Subscribe(channel)
	defer sub.Close()

	for {
		msg, err := sub.ReceiveMessage()
		if err != nil {
			return err
		}

		handler(msg.Payload)
	}
}

func LoadFromDB() error {
	db := GetDB()
	ctx, err := db.QueryContext(context.Background(), "SELECT * FROM tweets")
	if err != nil {
		return err
	}
	for ctx.Next() {
		var id int64

		err := ctx.Scan(&id)
		if err != nil {
			return err
		}
		idStr := strconv.FormatInt(id, 10)
		err = Set(idStr, idStr)
		if err != nil {
			return err
		}
	}
	return nil
}
