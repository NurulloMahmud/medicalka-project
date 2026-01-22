package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/NurulloMahmud/medicalka-project/utils"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func (m *Middleware) RateLimit(next http.Handler) http.Handler {
	if !m.cfg.Limiter.Enabbled {
		return next
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()

			for ip, client := range clients {
				if time.Since(client.lastSeen) >= time.Minute*3 {
					delete(clients, ip)
				}
			}

			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := realip.FromRequest(r)
		mu.Lock()

		if _, found := clients[ip]; !found {
			clients[ip] = &client{
				limiter: rate.NewLimiter(
					rate.Limit(m.cfg.Limiter.RPS),
					m.cfg.Limiter.Burst,
				),
			}
		}

		clients[ip].lastSeen = time.Now()

		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			utils.RateLimitExceeded(w, r)
			return
		}

		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
