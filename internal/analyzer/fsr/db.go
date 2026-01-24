package fsr

import (
	"errors"
	"net/netip"
	"time"

	"github.com/dgraph-io/badger/v4"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

// --- currentRecord ---

func (fsr *FileSendRatio) putCurrentRecord(rec currentRecord) error {
	recordBytes, err := rec.Marshal()
	if err != nil {
		return err
	}
	return fsr.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry(fsr.crKb.WithObject(rec).Build(), recordBytes)
		return txn.SetEntry(entry)
	})
}

func (fsr *FileSendRatio) getCurrentRecord(addr netip.Addr, path string) (currentRecord, error) {
	rec := currentRecord{
		Addr: addr,
		Path: path,
	}
	err := fsr.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(fsr.crKb.WithObject(rec).Build())
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return rec.Unmarshal(val)
		})
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return currentRecord{}, ErrRecordNotFound
		}
		return currentRecord{}, err
	}
	return rec, nil
}

func (fsr *FileSendRatio) clearCurrentRecords() error {
	return fsr.db.DropPrefix(fsr.crKb.Build())
}

// --- historicRecord ---

func (fsr *FileSendRatio) putHistoricRecords(rec []historicRecord) error {
	return fsr.db.Update(func(txn *badger.Txn) error {
		for _, v := range rec {
			recordBytes, err := v.Marshal()
			if err != nil {
				return err
			}
			entry := badger.NewEntry(fsr.hrKb.WithObject(v).Build(), recordBytes).WithTTL(time.Duration(fsr.cfg.RecordTTL))
			err = txn.SetEntry(entry)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (fsr *FileSendRatio) getMaxHistoricRecord() (historicRecord, error) {
	maxRec := historicRecord{
		Ratio: -1.,
	}
	err := fsr.db.View(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		opt.Prefix = fsr.hrKb.Build()
		it := txn.NewIterator(opt)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			var rec historicRecord
			err := item.Value(func(val []byte) error {
				return rec.Unmarshal(val)
			})
			if err != nil {
				return err
			}
			if rec.Ratio > maxRec.Ratio {
				maxRec = rec
			}
		}
		return nil
	})
	if err != nil {
		return historicRecord{}, err
	}
	return maxRec, nil
}

// --- fileSizeRecord ---

func (fsr *FileSendRatio) putFileSizeRecord(rec fileSizeRecord) error {
	recordBytes, err := rec.Marshal()
	if err != nil {
		return err
	}
	return fsr.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry(fsr.fsrKb.WithObject(rec).Build(), recordBytes).WithTTL(time.Duration(fsr.cfg.SizeInfoTTL))
		return txn.SetEntry(entry)
	})
}

func (fsr *FileSendRatio) getFileSizeRecord(path string) (fileSizeRecord, error) {
	rec := fileSizeRecord{
		Path: path,
	}
	err := fsr.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(fsr.fsrKb.WithObject(rec).Build())
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return rec.Unmarshal(val)
		})
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return fileSizeRecord{}, ErrRecordNotFound
		}
		return fileSizeRecord{}, err
	}
	return rec, nil
}

// --- ipRecord ---

func (fsr *FileSendRatio) putIPRecord(rec ipRecord) error {
	recordBytes, err := rec.Marshal()
	if err != nil {
		return err
	}
	return fsr.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry(fsr.fsrKb.WithObject(rec).Build(), recordBytes)
		return txn.SetEntry(entry)
	})
}

func (fsr *FileSendRatio) getAllIPRecords() ([]ipRecord, error) {
	recs := make([]ipRecord, 0)
	err := fsr.db.View(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		opt.Prefix = fsr.ipKb.Build()
		it := txn.NewIterator(opt)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			var rec ipRecord
			err := item.Value(func(val []byte) error {
				return rec.Unmarshal(val)
			})
			if err != nil {
				return err
			}
			recs = append(recs, rec)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return recs, nil
}

func (fsr *FileSendRatio) clearIPRecords() error {
	return fsr.db.DropPrefix(fsr.ipKb.Build())
}
