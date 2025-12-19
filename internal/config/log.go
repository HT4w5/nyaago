package config

type LogConfig struct {
	Access   string `json:"access"`
	LogLevel string `json:"log_level"`
	Json     bool   `json:"json"`
}
