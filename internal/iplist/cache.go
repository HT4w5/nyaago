package iplist

import (
	"net/netip"
	"time"

	"github.com/HT4w5/nyaago/pkg/dto"
	"github.com/allegro/bigcache/v3"
)

func (l *IPList) PutRule(e dto.Rule) error {
	if !e.Valid {
		return nil
	}
	// Set ExpiresOn
	e.ExpiresOn = time.Now().Add(l.cfg.IPList.RuleTTL.Duration)

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

func (l *IPList) GetRule(b netip.Addr) (dto.Rule, error) {
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
	if !record.Valid {
		return dto.Rule{}, nil
	}
	return record, nil
}

func (l *IPList) DelRule(b netip.Addr) error {
	_, err := l.cache.Get(b.String())
	if err != nil {
		if err == bigcache.ErrEntryNotFound {
			return nil
		} else {
			return err
		}
	}

	r, err := dto.Rule{}.MarshalBinary()
	if err != nil {
		return err
	}
	err = l.cache.Set(b.String(), r)
	if err != nil {
		return err
	}
	return nil
}
