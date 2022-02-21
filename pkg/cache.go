package pkg

import (
	"context"
	"github.com/go-redis/redis/v8"
	"strconv"
)

var cache *redis.Client
var ctx = context.Background()

func init() {
	cache = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func Set(key string, value string) error {
	return cache.Set(ctx, key, value, 0).Err()
}

func Get(key string) (string, error) {
	return cache.Get(ctx, key).Result()
}

func Del(key string) error {
	return cache.Del(ctx, key).Err()
}

func Flush() error {
	return cache.FlushDB(ctx).Err()
}

func Exists(key string) (int64, error) {
	return cache.Exists(ctx, key).Result()
}

func Publish(channel string, message string) error {
	return cache.Publish(ctx, channel, message).Err()
}

func Subscribe(channel string, handler func(string)) error {
	sub := cache.Subscribe(ctx, channel)
	defer sub.Close()

	for {
		msg, err := sub.ReceiveMessage(ctx)
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
