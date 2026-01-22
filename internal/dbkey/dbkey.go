package dbkey

import "encoding/binary"

type Prefix uint32

// Must return fixed-length slice for each type
// to avoid leaking
type Object interface {
	DBKey() []byte
}

type KeyBuilder struct {
	prefix   uint32
	segments []byte
}

func MakeKeyBuilder(prefix uint32) KeyBuilder {
	kb := KeyBuilder{
		prefix:   prefix,
		segments: make([]byte, 0),
	}
	binary.BigEndian.AppendUint32(kb.segments, prefix)
	return kb
}

func (kb KeyBuilder) WithObject(o Object) KeyBuilder {
	kb.segments = append(kb.segments, o.DBKey()...)
	return kb
}

func (kb KeyBuilder) Build() []byte {
	return kb.segments
}
