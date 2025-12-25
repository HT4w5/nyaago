package denylist

import (
	"net/netip"

	"github.com/allegro/bigcache/v3"
)

func (l *DenyList) putEntry(e Entry) error {
	err := l.cache.Set(e.Blame.String(), e.Marshal())
	if err != nil {
		return err
	}
	return nil
}

func (l *DenyList) getEntry(b netip.Addr) (Entry, error) {
	entryBytes, err := l.cache.Get(b.String())
	if err != nil {
		// Return empty Entry on not found
		if err == bigcache.ErrEntryNotFound {
			return Entry{}, nil
		}
		return Entry{}, err
	}
	var record Entry
	err = record.Unmarshal(entryBytes)
	if err != nil {
		return Entry{}, err
	}
	return record, nil
}
