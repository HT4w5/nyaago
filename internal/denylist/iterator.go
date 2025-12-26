package denylist

import (
	"iter"

	"github.com/HT4w5/nyaago/pkg/dto"
)

func (l *DenyList) Iterator() iter.Seq[dto.Rule] {
	return func(yield func(dto.Rule) bool) {
		it := l.cache.Iterator()
		for it.SetNext() {
			v, err := it.Value()
			if err != nil {
				continue
			}

			var rule dto.Rule
			if err := rule.UnmarshalBinary(v.Value()); err != nil {
				continue
			}

			if !yield(rule) {
				return
			}
		}
	}
}
