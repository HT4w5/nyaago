package config

type IngressConfig struct {
	Method string       `json:"method"`
	Format string       `json:"format"`
	Syslog SyslogConfig `json:"syslog"`
	Tail   TailConfig   `json:"tail"`
}

type SyslogConfig struct {
	Protocol   string `json:"protocol"`
	ListenAddr string `json:"listen_addr"`
}

type TailConfig struct {
	Path string `json:"path"`
	Poll bool   `json:"poll"`
}
