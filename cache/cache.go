package cache

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"strconv"
	"tw-bot/database"
)

var cache *redis.Client
var ctx = context.Background()

type Cache struct {
	client *redis.Client
}

func NewCache() *Cache {
	return &Cache{
		client: cache,
	}
}

func init() {
	cache = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func (c *Cache) SAdd(key string, value string) (int64, error) {
	return cache.SAdd(ctx, key, value).Result()
}
func (c *Cache) SPop(key string) (string, error) {
	return cache.SPop(ctx, key).Result()
}

func (c *Cache) SRem(key string, value string) {
	cache.SRem(ctx, key, value)
}
func (c *Cache) SRandMember(key string) (string, error) {
	return cache.SRandMember(ctx, key).Result()
}

func (c *Cache) SIsMember(key string, value string) (bool, error) {
	return cache.SIsMember(ctx, key, value).Result()
}

func (c *Cache) SMembers(key string) ([]string, error) {
	return cache.SMembers(ctx, key).Result()
}

func (c *Cache) Del(key string) error {
	return cache.Del(ctx, key).Err()
}

func (c *Cache) Flush() error {
	return cache.FlushDB(ctx).Err()
}

func (c *Cache) Exists(key string) (int64, error) {
	return cache.Exists(ctx, key).Result()
}

func (c *Cache) Publish(channel string, message string) error {
	return cache.Publish(ctx, channel, message).Err()
}

func (c *Cache) Subscribe(channel string) *redis.PubSub {

	var err error
	subPub := cache.Subscribe(ctx, channel)
	_, err = subPub.Receive(ctx)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	return subPub
}

func (c *Cache) LoadFromDB() error {
	db := database.GetDataBase()
	ctx, err := db.Sqlite.QueryContext(context.Background(), "SELECT * FROM tweets")
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
		_, err = c.SAdd("entity", idStr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Cache) HandlerSubscribe(pubSub *redis.PubSub) {
	ch := pubSub.Channel()
	for {
		msg, ok := <-ch
		if !ok {
			fmt.Println("receive message is wrong")
			break
		}

		fmt.Println(msg.Channel, msg.Payload)
	}
}

func (c *Cache) CloseSubscribe(pubSub *redis.PubSub) {
	err := pubSub.Close()
	if err != nil {
		return
	}
}
