package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// rateLimiters menyimpan map IP ke *rate.Limiter.
var rateLimiters = make(map[string]*rate.Limiter)
var mu sync.RWMutex // Mutex untuk melindungi akses ke map

// getLimiter membuat atau mendapatkan Limiter untuk IP tertentu.
func getLimiter(ip string) *rate.Limiter {
	mu.RLock() // Gunakan RLock untuk pembacaan
	if limiter, ok := rateLimiters[ip]; ok {
		mu.RUnlock()
		return limiter
	}
	mu.RUnlock()

	mu.Lock() // Gunakan Lock untuk penulisan
	defer mu.Unlock()

	if limiter, ok := rateLimiters[ip]; ok {
		return limiter
	}

	// 5 request per 5 detik, dengan burst 10
	newLimiter := rate.NewLimiter(rate.Every(5*time.Second/5), 10)
	rateLimiters[ip] = newLimiter

	return newLimiter
}

// LimitByIP membatasi request berdasarkan alamat IP klien.
func LimitByIP() gin.HandlerFunc {
	// Jalankan Goroutine untuk membersihkan Limiter yang tidak aktif
	go func() {
		for {
			time.Sleep(time.Minute * 5) // Bersihkan setiap 5 menit
			mu.Lock()
			// Logika pembersihan sederhana
			// Untuk produksi,  perlu logika pembersihan
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		// Dapatkan IP klien. Gunakan c.ClientIP() Gin.
		ip := c.ClientIP()
		limiter := getLimiter(ip)

		// Batasi 1 request per 1 detik (contoh)
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Batas permintaan per IP terlampaui. Mohon tunggu.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Cara Penggunaan di main.go:
/*
func main() {
    r := gin.Default()
    r.GET("/protected", middlewares.LimitByIP(), controller.ProtectedHandler)
    r.Run()
}
*/
