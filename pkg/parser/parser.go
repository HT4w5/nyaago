package parser

import (
	"fmt"

	"github.com/HT4w5/nyaago/pkg/dto"
)

type Parser interface {
	Parse(line []byte) (dto.Request, error)
}

func MakeParser(logFormat string) (Parser, error) {
	switch logFormat {
	case "nginxjson":
		return &NginxJSONParser{}, nil
	default:
		return nil, fmt.Errorf("unsupported log format: %s", logFormat)
	}
}
