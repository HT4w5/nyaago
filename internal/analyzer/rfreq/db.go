package rfreq

import (
	"errors"
	"time"

	"github.com/dgraph-io/badger/v4"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

func (rf *RequestFrequency) putRecords(recs []record) error {
	return rf.db.Update(func(txn *badger.Txn) error {
		for _, v := range recs {
			recordBytes, err := v.Marshal()
			if err != nil {
				return err
			}
			entry := badger.NewEntry(rf.kb.WithObject(v).Build(), recordBytes).WithTTL(time.Duration(rf.cfg.RecordTTL))
			err = txn.SetEntry(entry)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
