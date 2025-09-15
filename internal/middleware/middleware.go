package middleware

import (
	"github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
	redisstore "github.com/ulule/limiter/v3/drivers/store/redis"
)

type Middleware struct {
	limiter *limiter.Limiter
	apiKey  string
}

func NewMiddleware(redisClient *redis.Client, rateSpec string, apiKey string) (*Middleware, error) {
	rate, err := limiter.NewRateFromFormatted(rateSpec)
	if err != nil {
		return nil, err
	}
	store, err := redisstore.NewStoreWithOptions(redisClient, limiter.StoreOptions{
		Prefix: "rate_limiter",
	})
	if err != nil {
		return nil, err
	}
	return &Middleware{
		limiter: limiter.New(store, rate),
		apiKey:  apiKey,
	}, nil
}
