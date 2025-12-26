package aclfmt

import (
	"fmt"
	"io"

	"github.com/HT4w5/nyaago/pkg/dto"
)

type Formatter interface {
	Marshal(rules []dto.Rule, w io.Writer) error
	Info() string
}

func MakeFormatter(format string, info string) (Formatter, error) {
	switch format {
	case "nginx":
		return makeNginxFormatter(info), nil
	default:
		return nil, fmt.Errorf("unsupported formatter type %s", format)
	}
}
