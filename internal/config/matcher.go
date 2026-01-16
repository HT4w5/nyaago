package config

// A matcher for requests
type MatcherConfig struct {
	Client      *IPPrefix `json:"client"`
	Server      *IPPrefix `json:"server"`
	Method      *string   `json:"method"`
	URL         *Regexp   `json:"url"`
	Status      *int      `json:"status"`
	SentMin     *int64    `json:"sent_min"`
	SentMax     *int64    `json:"sent_max"`
	DurationMin *Duration `json:"duration_min"`
	DurationMax *Duration `json:"duration_max"`
	Host        *Regexp   `json:"host"`
	Agent       *Regexp   `json:"agent"`
}
