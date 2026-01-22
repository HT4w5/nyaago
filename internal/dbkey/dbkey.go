package dbkey

type Prefix byte

const (
	LeakyBucket Prefix = iota
	FileSendRatio
	RequestFrequency
)

// Must return fixed-length slice for each type
// for best performance
type Object interface {
	DBKey() []byte
}

type KeyBuilder struct {
	bytes []byte
}

func (kb KeyBuilder) WithPrefix(p Prefix) KeyBuilder {
	kb.bytes = append(kb.bytes, byte(p))
	return kb
}

func (kb KeyBuilder) WithObject(o Object) KeyBuilder {
	kb.bytes = append(kb.bytes, o.DBKey()...)
	return kb
}

func (kb KeyBuilder) Build() []byte {
	return kb.bytes
}
