package analyzer

import (
	"iter"

	"github.com/HT4w5/nyaago/pkg/dto"
)

func (a *Analyzer) Iterator() iter.Seq[dto.Record] {
	return func(yield func(dto.Record) bool) {
		it := a.cache.Iterator()
		for it.SetNext() {
			v, err := it.Value()
			if err != nil {
				continue
			}

			var r dto.Record
			if err := r.Unmarshal(v.Value()); err != nil {
				continue
			}

			if !yield(r) {
				return
			}
		}
	}
}
