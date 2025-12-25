package analyzer

import (
	"net/netip"

	"github.com/allegro/bigcache/v3"
)

func (a *Analyzer) putRecord(r Record) error {
	err := a.cache.Set(r.Addr.String(), r.Marshal())
	if err != nil {
		return err
	}
	return nil
}

func (a *Analyzer) getRecord(addr netip.Addr) (Record, error) {
	recordBytes, err := a.cache.Get(addr.String())
	if err != nil {
		// Return empty Record on not found
		if err == bigcache.ErrEntryNotFound {
			return Record{}, nil
		}
		return Record{}, err
	}
	var record Record
	err = record.Unmarshal(recordBytes)
	if err != nil {
		return Record{}, err
	}
	return record, nil
}
