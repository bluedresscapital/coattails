package wardrobe

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
)

var cache redis.Conn

func InitCache(host string) {
	// Initialize the redis connection to a redis instance running on your local machine
	conn, err := redis.DialURL(fmt.Sprintf("redis://%s", host))
	if err != nil {
		panic(err)
	}
	// Assign the connection to the package level `cache` variable
	cache = conn
}
