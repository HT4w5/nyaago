package config

import (
	"fmt"
	"net/netip"

	"github.com/HT4w5/nyaago/pkg/dto"
)

type RequestFilter struct {
	Prefix   netip.Prefix `json:"prefix"`
	URLRegex RegexWrapper `json:"url_regex"`
}

/*
Functions that apply filters on incoming requests.
*/

func (f RequestFilter) Match(request dto.Request) bool {
	if f.Prefix.IsValid() && !f.Prefix.Contains(request.Client) {
		return false
	}

	if f.URLRegex.IsValid() && !f.URLRegex.MatchString(request.URL) {
		return false
	}

	return true
}

func (f RequestFilter) String() string {
	prefixStr := "nil"
	urlRegexStr := "nil"
	if f.Prefix.IsValid() {
		prefixStr = f.Prefix.String()
	}
	if f.URLRegex.isValid {
		urlRegexStr = f.URLRegex.String()
	}
	return fmt.Sprintf("Prefix: %s URLRegex: %s", prefixStr, urlRegexStr)
}
