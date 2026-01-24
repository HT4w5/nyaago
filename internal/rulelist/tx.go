package rulelist

import (
	"github.com/HT4w5/nyaago/internal/dbkey"
	"github.com/HT4w5/nyaago/pkg/dto"
	"github.com/dgraph-io/badger/v4"
)

type Tx struct {
	tx *badger.Txn
	kb dbkey.KeyBuilder
}

func (rl *RuleList) BeginTx() *Tx {
	return &Tx{
		tx: rl.db.NewTransaction(true),
		kb: rl.kb,
	}
}

func (tx *Tx) Commit() error {
	return tx.tx.Commit()
}

func (tx *Tx) Discard() {
	tx.tx.Discard()
}

func (tx *Tx) PutRule(rule dto.Rule) error {
	entryBytes, err := rule.Marshal()
	if err != nil {
		return err
	}

	entry := badger.NewEntry(tx.kb.WithObject(rule).Build(), entryBytes)
	entry.ExpiresAt = uint64(rule.ExpiresAt.Unix())
	return tx.tx.SetEntry(entry)
}
