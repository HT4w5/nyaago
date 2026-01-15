package lbucket

import (
	"errors"
	"net/netip"

	"github.com/dgraph-io/badger/v4"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

func (lb *LeakyBucket) putRecord(rec record) error {
	recordBytes, err := rec.Marshal()
	if err != nil {
		return err
	}
	return lb.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry(lb.kb.WithObject(rec).Build(), recordBytes)
		return txn.SetEntry(entry)
	})
}

func (lb *LeakyBucket) getRecord(addr netip.Addr) (record, error) {
	rec := record{
		Addr: addr,
	}
	err := lb.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(lb.kb.WithObject(rec).Build())
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return rec.Unmarshal(val)
		})
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return record{}, ErrRecordNotFound
		}
		return record{}, err
	}
	return rec, nil
}
