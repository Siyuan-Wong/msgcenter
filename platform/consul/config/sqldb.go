package config

type SqlDb struct {
	Id      string   `json:"id"`
	Schema  string   `json:"schema"`
	Hosts   []string `json:"hosts"`
	Account `json:"account"`
}

type Account struct {
	Encrypt  bool   `json:"encrypt"`
	User     string `json:"user"`
	Password string `json:"password"`
}
