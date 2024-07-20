package utils

import "github.com/go-redis/redis"

var RedisClient *redis.Client

func CreateRedisClient() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "192.168.1.50:6379",
		Password: "",
		DB:       0,
	})
}
