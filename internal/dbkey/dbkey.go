package dbkey

type Tag byte

const (
	RuleListTag Tag = iota
	LeakyBucketTag
)

// Must return fixed-length slice for each type
// to avoid leaking
type Object interface {
	DBKey() []byte
}

type KeyBuilder struct {
	segments []byte
}

func (kb *KeyBuilder) WithTag(s Tag) *KeyBuilder {
	kb.segments = append(kb.segments, byte(s))
	return kb
}

func (kb *KeyBuilder) WithObject(o Object) *KeyBuilder {
	kb.segments = append(kb.segments, o.DBKey()...)
	return kb
}

func (kb *KeyBuilder) Build() []byte {
	return kb.segments
}
