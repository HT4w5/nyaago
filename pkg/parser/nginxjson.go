package parser

import (
	"encoding/json"
	"net/netip"
	"time"

	"github.com/HT4w5/nyaago/pkg/dto"
)

type nginxJSONLogEntry struct {
	Timestamp     float64 `json:"timestamp"` // $msec format: "1734345934.123"
	RemoteAddr    string  `json:"remote_addr"`
	RequestMethod string  `json:"request_method"`
	RequestURI    string  `json:"request_uri"`
	Status        string  `json:"status"`
	BodyBytesSent int64   `json:"body_bytes_sent"`
}

type NginxJSONParser struct {
}

func (p *NginxJSONParser) Parse(line []byte) (dto.Request, error) {
	var logEntry nginxJSONLogEntry
	err := json.Unmarshal(line, &logEntry)
	if err != nil {
		return dto.Request{}, err
	}

	e := dto.Request{
		URL:      logEntry.RequestURI,
		BodySent: logEntry.BodyBytesSent,
		Time:     time.Unix(int64(logEntry.Timestamp), 0), // Drop milisec precision
	}

	e.Client, err = netip.ParseAddr(logEntry.RemoteAddr)
	if err != nil {
		return dto.Request{}, err
	}

	return e, nil
}
