package meta

import "fmt"

var (
	Name       = "Nyaago"
	BuildDate  string
	CommitHash string
	Version    string
	Platform   string
	GoVersion  string
)

type buildInfo struct{}

func (b buildInfo) Lines() []string {
	return []string{
		fmt.Sprintf(
			"%s %s (%s %s)",
			Name,
			Version,
			GoVersion,
			Platform,
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
