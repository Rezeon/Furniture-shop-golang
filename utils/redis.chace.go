package utils

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client

func InitRedis() {

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Username: os.Getenv("REDIS_USERNAME"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	maxRetries := 5
	retryDelay := 4 * time.Second

	for i := 0; i < maxRetries; i++ {
		_, err := RedisClient.Ping(ctx).Result()

		if err == nil {
			fmt.Println("Koneksi Redis berhasil.")
			return
		}
		if i < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}

	log.Fatalf("Aplikasi mati: Gagal terhubung ke Redis setelah %d percobaan.", maxRetries)
}
