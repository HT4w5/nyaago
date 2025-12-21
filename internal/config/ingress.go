package config

type IngressConfig struct {
	Method string       `json:"method"`
	Format string       `json:"format"`
	Syslog SyslogConfig `json:"syslog"`
	Tail   TailConfig   `json:"tail"`
}

type SyslogConfig struct {
	Transport  string `json:"transport"`
	ListenAddr string `json:"listen_addr"`
}

type TailConfig struct {
	Path string `json:"path"`
	Poll bool   `json:"poll"`
}
