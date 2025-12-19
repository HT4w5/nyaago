package config

type EgressConfig struct {
	Path     string         `json:"path"`
	Type     string         `json:"type"`
	PostExec PostExecConfig `json:"post_exec"`
}

type PostExecConfig struct {
	Cmd  string   `json:"cmd"`
	Cwd  string   `json:"cwd"`
	Args []string `json:"args"`
}
