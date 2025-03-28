package config

import "fmt"

type SqlDb struct {
	Host            string `json:"host"`
	Port            string `json:"port"`
	User            string `json:"user"`
	Password        string `json:"password"`
	Database        string `json:"database"`
	Schema          string `json:"schema"`
	MaxIdleConns    int    `json:"max_idle_conns"`
	MaxOpenConns    int    `json:"max_open_conns"`
	ConnMaxLifetime int    `json:"conn_max_lifetime"`
	ConnMaxIdleTime int    `json:"conn_max_idle_time"`
	SSLMode         string `json:"ssl_mode"`
}

func (c *SqlDb) Dsn() string {
	sslMode := "disable"
	if c.SSLMode != "" {
		sslMode = c.SSLMode
	}

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s search_path=%s sslmode=%s",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.Database,
		c.Schema,
		sslMode,
	)
}
