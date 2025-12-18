package meta

import "fmt"

const (
	Name       = "Nyaago"
	BuildDate  = ""
	CommitHash = ""
	Version    = "v2025.12-alpha"
)

type buildInfo struct{}

func (b buildInfo) Lines() []string {
	return []string{
		fmt.Sprintf(
			"%s %s",
			Name,
			Version,
		),
		fmt.Sprintf(
			"Build date: %s",
			BuildDate,
		),
		fmt.Sprintf(
			"Commit hash: %s",
			CommitHash,
		),
	}
}

func (b buildInfo) Width() int {
	return getWidth(b.Lines())
}

func (b buildInfo) Config() fieldConfig {
	return fieldConfig{
		Alignment: fieldAlignLeft,
	}
}

func getBuildInfo() field {
	return buildInfo{}
}
