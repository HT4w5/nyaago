package config

type IngressConfig struct {
	Path string `json:"path"`
	Type string `json:"type"`
	Poll bool   `json:"poll"`
}
