package rulelist

import (
	"net/netip"

	"github.com/HT4w5/nyaago/pkg/dto"
	"github.com/dgraph-io/badger/v4"
)

func (l *RuleList) PutRule(rule dto.Rule) error {
	entryBytes, err := rule.Marshal()
	if err != nil {
		return err
	}
	return l.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry(l.kb.WithObject(rule).Build(), entryBytes)
		entry.ExpiresAt = uint64(rule.ExpiresAt.Unix())
		return txn.SetEntry(entry)
	})
}

func (l *RuleList) GetRule(prefix netip.Prefix) (dto.Rule, error) {
	rule := dto.Rule{
		Prefix: prefix,
	}
	err := l.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(l.kb.WithObject(rule).Build())
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return rule.Unmarshal(val)
		})
	})
	if err != nil {
		return dto.Rule{}, err
	}
	return rule, nil
}

func (l *RuleList) DelRule(prefix netip.Prefix) error {
	rule := dto.Rule{
		Prefix: prefix,
	}

	return l.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(l.kb.WithObject(rule).Build())
	})
}
