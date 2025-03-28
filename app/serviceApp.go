package app

import (
	"msgcenter/config"
	"msgcenter/ent"
	"msgcenter/platform/consul"
	"sync"
)

var (
	instance *ServiceApp
	once     sync.Once
)

type ServiceApp struct {
	LocalConfig *config.Config
	DbClient    *ent.Client
	Consul      *consul.Client
}

func GetService(localConfig *config.Config, dbClient *ent.Client, consul *consul.Client) *ServiceApp {
	once.Do(func() {
		instance = &ServiceApp{
			LocalConfig: localConfig,
			DbClient:    dbClient,
			Consul:      consul,
		}
	})
	return instance
}

func SERVICE() *ServiceApp {
	return instance
}
