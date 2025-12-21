package aclfmt

import (
	"fmt"
	"io"

	"go4.org/netipx"
)

type Formatter interface {
	Marshal(ipset *netipx.IPSet, w io.Writer) error
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
