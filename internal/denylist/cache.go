package denylist

import (
	"net/netip"

	"github.com/HT4w5/nyaago/pkg/dto"
	"github.com/allegro/bigcache/v3"
)

func (l *DenyList) putRule(e dto.Rule) error {
	entryBytes, err := e.MarshalBinary()
	if err != nil {
		return err
	}
	err = l.cache.Set(e.Blame.String(), entryBytes)
	if err != nil {
		return err
	}
	return nil
}

func (l *DenyList) getRule(b netip.Addr) (dto.Rule, error) {
	entryBytes, err := l.cache.Get(b.String())
	if err != nil {
		// Return empty dto.Rule on not found
		if err == bigcache.ErrEntryNotFound {
			return dto.Rule{}, nil
		}
		return dto.Rule{}, err
	}
	var record dto.Rule
	err = record.UnmarshalBinary(entryBytes)
	if err != nil {
		return dto.Rule{}, err
	}
	return record, nil
}
