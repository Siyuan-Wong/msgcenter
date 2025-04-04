package app

import (
	"github.com/redis/go-redis/v9"
	"msgcenter/config"
	"msgcenter/platform/consul"
	"msgcenter/platform/ent/gen"
	"sync"
)

var (
	instance *ServiceApp
	once     sync.Once
)

type ServiceApp struct {
	LocalConfig *config.Config
	DbClient    *gen.Client
	Consul      *consul.Client
	RedisClient *redis.Client
}

func GetService(localConfig *config.Config, dbClient *gen.Client, consul *consul.Client, redisClient *redis.Client) *ServiceApp {
	once.Do(func() {
		instance = &ServiceApp{
			LocalConfig: localConfig,
			DbClient:    dbClient,
			Consul:      consul,
			RedisClient: redisClient,
		}
	})
	return instance
}

func SERVICE() *ServiceApp {
	return instance
}
