package meta

import (
	"fmt"
	"strings"
)

const separator = "─"

type fieldAlignment int

const (
	fieldAlignLeft fieldAlignment = iota
	fieldAlignCenter
	fieldAlignRight
)

type field interface {
	Lines() []string
	Width() int
	Config() fieldConfig
}

type fieldConfig struct {
	Alignment fieldAlignment
}

func GetMetadataMultiline() string {
	fields := []field{
		getBanner(),
		getBuildInfo(),
		getMotd(),
	}

	// Get separator length
	spLen := 0
	for _, v := range fields {
		l := v.Width()
		if l > spLen {
			spLen = l
		}
	}

	var sb strings.Builder
	// Omit error check
	sep := strings.Repeat(separator, spLen)
	sb.WriteString(sep)
	sb.WriteRune('\n')

	for _, v := range fields {
		cfg := v.Config()
		padding := ""
		switch cfg.Alignment {
		case fieldAlignLeft:
			padding = ""
		case fieldAlignCenter:
			padding = strings.Repeat(" ", (spLen-v.Width())/2)
		case fieldAlignRight:
			padding = strings.Repeat(" ", spLen-v.Width())
		}
		for _, l := range v.Lines() {
			sb.WriteString(padding)
			sb.WriteString(l)
			sb.WriteRune('\n')
		}

		sb.WriteString(sep)
		sb.WriteRune('\n')
	}

	return sb.String()
}

func GetMetadataSingleLine() string {
	return fmt.Sprintf(
		"%s %s %s",
		Name,
		Version,
		CommitHash,
	)
}
