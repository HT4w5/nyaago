package iplist

import (
	"iter"
)

func (l *IPList) Iterator() iter.Seq[IPEntry] {
	return func(yield func(IPEntry) bool) {
		it := l.cache.Iterator()
		for it.SetNext() {
			v, err := it.Value()
			if err != nil {
				continue
			}

			var e IPEntry
			if err := e.UnmarshalBinary(v.Value()); err != nil {
				continue
			}

			if !yield(e) {
				return
			}
		}
	}
}
