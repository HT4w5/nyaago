package analyzer

import (
	"net/netip"

	"github.com/HT4w5/nyaago/pkg/dto"
)

func (a *Analyzer) putRecord(r dto.Record) error {
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

func (a *Analyzer) getRecord(addr netip.Addr) (dto.Record, error) {
	recordBytes, err := a.cache.Get(addr.String())
	if err != nil {
		return dto.Record{}, err
	}
	var record dto.Record
	err = record.Unmarshal(recordBytes)
	if err != nil {
		return dto.Record{}, err
	}
	return record, nil
}

func (a *Analyzer) Len() int {
	return a.cache.Len()
}
