package config

import "time"

type Redis struct {
	Addr         string        `json:"addr"`
	Password     string        `json:"password"`
	DB           int           `json:"db"`
	PoolSize     int           `json:"pool_size"`
	MinIdleConns int           `json:"min_idle_conns"`
	MaxIdleConns int           `json:"max_idle_conns"`
	PoolTimeout  time.Duration `json:"idle_timeout"`
}
