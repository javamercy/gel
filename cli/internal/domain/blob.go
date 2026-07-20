package domain

import "slices"

// Blob represents immutable file content stored as a blob object.
type Blob struct {
	content []byte
}

// NewBlob returns a Blob containing a defensive copy of content.
func NewBlob(content []byte) *Blob {
	return &Blob{
		content: slices.Clone(content),
	}
}

// Type returns ObjectTypeBlob.
func (b *Blob) Type() ObjectType {
	return ObjectTypeBlob
}

func (b *Blob) isObject() {}

func (b *Blob) Content() []byte {
	return slices.Clone(b.content)
}
