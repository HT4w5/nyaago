package config

type EgressConfig struct {
	Path     string           `json:"path"`
	Format   string           `json:"format"`
	PostExec []PostExecConfig `json:"post_exec"`
}

type PostExecConfig struct {
	Tag  string   `json:"tag"`
	Cmd  string   `json:"cmd"`
	Cwd  string   `json:"cwd"`
	Args []string `json:"args"`
}
