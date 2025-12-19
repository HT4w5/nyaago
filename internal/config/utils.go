package config

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"
)

type Duration struct {
	time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		return err
	default:
		return fmt.Errorf("invalid duration type: %T", v)
	}
}

type RegexWrapper struct {
	*regexp.Regexp
}

func (r *RegexWrapper) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	re, err := regexp.Compile(s)
	if err != nil {
		return err
	}

	r.Regexp = re
	return nil
}
