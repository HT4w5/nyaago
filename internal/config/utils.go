package config

import (
	"net/netip"
	"regexp"
	"strings"
	"time"

	"github.com/docker/go-units"
)

type Duration time.Duration

func (d *Duration) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)

	du, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*d = Duration(du)
	return nil
}

type Regexp struct {
	*regexp.Regexp
}

func (r *Regexp) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)

	re, err := regexp.Compile(s)
	if err != nil {
		return err
	}

	r.Regexp = re
	return nil
}

type ByteSize int64

func (b *ByteSize) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)

	res, err := units.FromHumanSize(s)
	if err != nil {
		return err
	}

	*b = ByteSize(res)
	return nil
}

type IPPrefix struct {
	netip.Prefix
}

func (p *IPPrefix) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)

	var err error
	p.Prefix, err = netip.ParsePrefix(s)
	return err
}
