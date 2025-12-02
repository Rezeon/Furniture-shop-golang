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
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
		log.Printf("Peringatan: REDIS_ADDR tidak disetel, menggunakan default: %s", redisAddr)
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	maxRetries := 5
	retryDelay := 4 * time.Second

	for i := 0; i < maxRetries; i++ {
		_, err := RedisClient.Ping(ctx).Result()

		if err == nil {
			fmt.Println("Koneksi Redis berhasil.")
			return // KELUAR DARI FUNGSI JIKA BERHASIL
		}

		log.Printf("Gagal terhubung ke Redis (Percobaan %d/%d ke %s): %v", i+1, maxRetries, redisAddr, err)

		if i < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}

	log.Fatalf("Aplikasi mati: Gagal terhubung ke Redis setelah %d percobaan.", maxRetries)
}
