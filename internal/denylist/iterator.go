package denylist

import (
	"iter"
)

func (l *DenyList) Iterator() iter.Seq[Entry] {
	return func(yield func(Entry) bool) {
		it := l.cache.Iterator()
		for it.SetNext() {
			v, err := it.Value()
			if err != nil {
				l.logger.Error("failed to get entry from cache", "error", err)
				continue
			}

			var entry Entry
			if err := entry.Unmarshal(v.Value()); err != nil {
				l.logger.Error("failed to unmarshal entry", "error", err)
				continue
			}

			if !yield(entry) {
				return
			}
		}
	}
}
