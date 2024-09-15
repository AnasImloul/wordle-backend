package src

import (
	"context"
	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

var rdb = redis.NewClient(&redis.Options{
	Addr:     "redis:6379", // Use the service name 'redis'
	Password: "",           // no password set
	DB:       0,            // use default DB
})
