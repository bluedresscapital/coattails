package wardrobe

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"log"
)

var (
	cache *redis.Client
)

func InitCache(host string) {
	// Initialize the redis connection to a redis instance
	cache = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:6379", host),
		Password: "",
		DB:       0,
	})
	if cache == nil {
		log.Fatal("Error in connecting to redis server")
	}
	err := cache.Ping().Err()
	if err != nil {
		log.Fatal("Error in connecting to redis server")
	}
}

func Publish(channel string, message []byte) error {
	return cache.Publish(channel, message).Err()
}

func Sub(channel string) *redis.PubSub {
	return cache.Subscribe(channel)
}

func Unsub(sub *redis.PubSub, channel string) error {
	return sub.Unsubscribe(channel)
}
