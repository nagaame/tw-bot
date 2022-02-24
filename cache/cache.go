package cache

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"tw-bot/database"
	"tw-bot/tool"
)

var cache *redis.Client
var ctx = context.Background()

type Cache struct {
	client *redis.Client
}

func NewRedisCache() *Cache {
	return &Cache{
		client: cache,
	}
}

func init() {
	cache = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx := context.Background()
	err := cache.Ping(ctx).Err()
	if err != nil {
		panic(err)
	}
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

func (c *Cache) Set(key string, value string) (string, error) {
	return cache.Set(ctx, key, value, 0).Result()
}

func (c *Cache) Get(key string) (string, error) {
	return cache.Get(ctx, key).Result()
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

func (c *Cache) XAdd(key string, value interface{}) (string, error) {
	return cache.XAdd(ctx, &redis.XAddArgs{
		Stream: key,
		Values: map[string]interface{}{
			"tweet": value,
		},
	}).Result()
}

func (c *Cache) XRange(key string, start string, end string) ([]redis.XMessage, error) {
	return cache.XRange(ctx, key, start, end).Result()
}

func (c *Cache) XDel(key string, value string) (int64, error) {
	return cache.XDel(ctx, key, value).Result()
}

func (c *Cache) XRead(key string) ([]redis.XStream, error) {
	return cache.XRead(ctx, &redis.XReadArgs{
		Streams: []string{key},
		Count:   1,
		Block:   0,
	}).Result()
}

func (c *Cache) XGroupCreate(key string, group string) (string, error) {
	return cache.XGroupCreate(ctx, key, group, "0").Result()
}

func (c *Cache) XReadGroup(key string, group string, consumer string, count int64) ([]redis.XStream, error) {
	return cache.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  []string{key, ">"},
		Count:    count,
		Block:    0,
		NoAck:    false,
	}).Result()
}

func (c *Cache) XReadBlock(key string) ([]redis.XStream, error) {
	return cache.XRead(ctx, &redis.XReadArgs{
		Streams: []string{key},
		Count:   1,
		Block:   -1,
	}).Result()
}

func (c *Cache) XAck(key string, value string) (int64, error) {
	return cache.XAck(ctx, key, value).Result()
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
		idStr := tool.IntToString(id)
		_, err = c.SAdd("tweets", idStr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Cache) HandlerSubscribe(pubSub *redis.PubSub, handler func(string, string)) {
	ch := pubSub.Channel()
	for {
		msg, ok := <-ch
		if !ok {
			fmt.Println("receive message is wrong")
			break
		}
		fmt.Println(msg.Channel, msg.Payload)
		handler(msg.Channel, msg.Payload)
	}
}

func (c *Cache) CloseSubscribe(pubSub *redis.PubSub) {
	err := pubSub.Close()
	if err != nil {
		return
	}
}
