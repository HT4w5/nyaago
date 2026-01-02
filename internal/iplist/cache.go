package iplist

import (
	"net/netip"
	"time"

	"github.com/allegro/bigcache/v3"
)

func (l *IPList) PutEntry(e IPEntry) error {
	if !e.Valid {
		return nil
	}
	// Set ExpiresOn
	e.ExpiresOn = time.Now().Add(l.cfg.IPList.EntryTTL.Duration)

	entryBytes, err := e.Marshal()
	if err != nil {
		return err
	}
	err = l.cache.Set(e.Addr.String(), entryBytes)
	if err != nil {
		return err
	}
	return nil
}

func (l *IPList) GetEntry(a netip.Addr) (IPEntry, error) {
	entryBytes, err := l.cache.Get(a.String())
	if err != nil {
		// Return empty IPEntry on not found
		if err == bigcache.ErrEntryNotFound {
			return IPEntry{}, nil
		}
		return IPEntry{}, err
	}
	var record IPEntry
	err = record.Unmarshal(entryBytes)
	if err != nil {
		return IPEntry{}, err
	}
	if !record.Valid {
		return IPEntry{}, nil
	}
	return record, nil
}

func (l *IPList) DelEntry(a netip.Addr) error {
	_, err := l.cache.Get(a.String())
	if err != nil {
		if err == bigcache.ErrEntryNotFound {
			return nil
		} else {
			return err
		}
	}

	e := IPEntry{}
	r, err := e.Marshal()
	if err != nil {
		return err
	}
	err = l.cache.Set(a.String(), r)
	if err != nil {
		return err
	}
	return nil
}
