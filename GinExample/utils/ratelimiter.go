package utils

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type Client struct {
	Limiter  *rate.Limiter
	LastSeen time.Time
}

type IPRateLimiter struct {
	Clients map[string]*Client
	Mu      sync.Mutex
}

func NewIPRateLimiter() *IPRateLimiter {
	return &IPRateLimiter{
		Clients: make(map[string]*Client),
	}
}

func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.Mu.Lock()
	defer i.Mu.Unlock()

	client, exists := i.Clients[ip]
	if !exists {
		limiter := rate.NewLimiter(rate.Every(time.Minute), 10)
		i.Clients[ip] = &Client{
			Limiter:  limiter,
			LastSeen: time.Now(),
		}
		return limiter
	}

	client.LastSeen = time.Now()
	return client.Limiter
}