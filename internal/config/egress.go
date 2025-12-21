package config

type EgressConfig struct {
	Path     string           `json:"path"`
	Type     string           `json:"type"`
	PostExec []PostExecConfig `json:"post_exec"`
}

type PostExecConfig struct {
	Tag  string   `json:"tag"`
	Cmd  string   `json:"cmd"`
	Cwd  string   `json:"cwd"`
	Args []string `json:"args"`
}
