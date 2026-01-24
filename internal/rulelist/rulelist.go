package rulelist

import (
	"github.com/HT4w5/nyaago/internal/config"
	"github.com/HT4w5/nyaago/internal/dbkey"
	"github.com/HT4w5/nyaago/pkg/dto"
	"github.com/dgraph-io/badger/v4"
)

type RuleList struct {
	cfg *config.Config
	db  *badger.DB
	kb  dbkey.KeyBuilder
}

func MakeRuleList(cfg *config.Config, db *badger.DB) (*RuleList, error) {
	l := &RuleList{
		cfg: cfg,
		db:  db,
		kb:  dbkey.KeyBuilder{}.WithPrefix(dbkey.RuleList),
	}

	return l, nil
}

func (l *RuleList) ListRules() ([]dto.Rule, error) {
	rules := make([]dto.Rule, 0)
	err := l.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		keyPrefix := l.kb.Build()
		for it.Seek(keyPrefix); it.ValidForPrefix(keyPrefix); it.Next() {
			item := it.Item()
			var rule dto.Rule
			err := item.Value(func(val []byte) error {
				return rule.Unmarshal(val)
			})
			if err != nil {
				return err
			}
			rules = append(rules, rule)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return rules, nil
}
