package cache

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"tw-bot/database"
	"tw-bot/tool"
)

var (
	client *redis.Client
	ctx    = context.Background()
	cache  *Cache
)

type Cache struct {
	client *redis.Client
}

func NewRedisCache() *Cache {

	if cache != nil {
		cache.client = client
		return cache
	}

	cache = new(Cache)
	cache.client = client
	return cache
}

func init() {
	client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx := context.Background()
	err := client.Ping(ctx).Err()
	if err != nil {
		panic(err)
	}
}

func (c *Cache) SAdd(key string, value ...interface{}) (int64, error) {
	return client.SAdd(ctx, key, value...).Result()
}
func (c *Cache) SPop(key string) (string, error) {
	return client.SPop(ctx, key).Result()
}

// SDiff redis set 差集
func (c *Cache) SDiff(keys ...string) ([]string, error) {
	return client.SDiff(ctx, keys...).Result()
}

// SScan redis set 迭代
func (c *Cache) SScan(key string, cursor uint64, match string, count int64) ([]string, uint64, error) {
	return client.SScan(ctx, key, cursor, match, count).Result()
}

func (c *Cache) SRem(key string, value string) {
	client.SRem(ctx, key, value)
}
func (c *Cache) SRandMember(key string) (string, error) {
	return client.SRandMember(ctx, key).Result()
}

func (c *Cache) SIsMember(key string, value string) (bool, error) {
	return client.SIsMember(ctx, key, value).Result()
}

func (c *Cache) SMembers(key string) ([]string, error) {
	return client.SMembers(ctx, key).Result()
}

func (c *Cache) Set(key string, value string) (string, error) {
	return client.Set(ctx, key, value, 0).Result()
}

func (c *Cache) Get(key string) (string, error) {
	return client.Get(ctx, key).Result()
}

func (c *Cache) Del(key string) error {
	return client.Del(ctx, key).Err()
}

func (c *Cache) Flush() error {
	return client.FlushDB(ctx).Err()
}

func (c *Cache) Exists(key string) (int64, error) {
	return client.Exists(ctx, key).Result()
}

func (c *Cache) Publish(channel string, message string) error {
	return client.Publish(ctx, channel, message).Err()
}

func (c *Cache) Subscribe(channel string) *redis.PubSub {

	var err error
	subPub := client.Subscribe(ctx, channel)
	_, err = subPub.Receive(ctx)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	return subPub
}

func (c *Cache) XAdd(key string, value interface{}) (string, error) {
	return client.XAdd(ctx, &redis.XAddArgs{
		Stream: key,
		Values: map[string]interface{}{
			"tweet": value,
		},
	}).Result()
}

func (c *Cache) XRange(key string, start string, end string) ([]redis.XMessage, error) {
	return client.XRange(ctx, key, start, end).Result()
}

func (c *Cache) XDel(key string, value string) (int64, error) {
	return client.XDel(ctx, key, value).Result()
}

func (c *Cache) XRead(key string) ([]redis.XStream, error) {
	return client.XRead(ctx, &redis.XReadArgs{
		Streams: []string{key},
		Count:   1,
		Block:   0,
	}).Result()
}

func (c *Cache) XGroupCreate(key string, group string) (string, error) {
	return client.XGroupCreate(ctx, key, group, "0").Result()
}

func (c *Cache) XReadGroup(key string, group string, consumer string, count int64) ([]redis.XStream, error) {
	return client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  []string{key, ">"},
		Count:    count,
		Block:    0,
		NoAck:    false,
	}).Result()
}

func (c *Cache) XReadBlock(key string) ([]redis.XStream, error) {
	return client.XRead(ctx, &redis.XReadArgs{
		Streams: []string{key},
		Count:   1,
		Block:   -1,
	}).Result()
}

func (c *Cache) XAck(key string, value string) (int64, error) {
	return client.XAck(ctx, key, value).Result()
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
