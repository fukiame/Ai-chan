package cache

import (
	"context"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/dgraph-io/ristretto"
	"github.com/fukiame/Ai-chan/ai/config"
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/marshaler"
	redis_store "github.com/eko/gocache/store/redis/v4"
	ristretto_store "github.com/eko/gocache/store/ristretto/v4"
	"github.com/redis/go-redis/v9"
)

var (
	Context = context.Background()
	Marshal *marshaler.Marshaler
	Manager *cache.ChainCache[any]
)

type AdminCache struct {
	ChatId   int64
	UserInfo []gotgbot.MergedChatMember
	Cached   bool
}

// InitCache initializes the cache.
func InitCache() {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddress,
		Password: config.RedisPassword, // no password set
		DB:       config.RedisDB,       // use default DB
	})
	ristrettoCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1000,
		MaxCost:     100,
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}

	// initialize cache manager
	redisStore := redis_store.NewRedis(redisClient)
	ristrettoStore := ristretto_store.NewRistretto(ristrettoCache)
	cacheManager := cache.NewChain(cache.New[any](ristrettoStore), cache.New[any](redisStore))

	// Initializes marshaler
	Marshal = marshaler.New(cacheManager)
}
