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

func MakeFormatter(fType string, info string) (Formatter, error) {
	switch fType {
	case "nginx":
		return makeNginxFormatter(info), nil
	default:
		return nil, fmt.Errorf("unsupported formatter type %s", fType)
	}
}
