package analyzer

import (
	"net/netip"
)

func (a *Analyzer) putRecord(r Record) error {
	recordBytes, err := r.Marshal()
	if err != nil {
		return err
	}
	err = a.cache.Set(r.Addr.String(), recordBytes)
	if err != nil {
		return err
	}
	return nil
}

func (a *Analyzer) getRecord(addr netip.Addr) (Record, error) {
	recordBytes, err := a.cache.Get(addr.String())
	if err != nil {
		return Record{}, err
	}
	var record Record
	err = record.Unmarshal(recordBytes)
	if err != nil {
		return Record{}, err
	}
	return record, nil
}
