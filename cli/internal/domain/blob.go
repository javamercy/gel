package domain

import "slices"

// Blob represents immutable file content stored as a blob object.
type Blob struct {
	body []byte
}

// NewBlob returns a Blob containing a defensive copy of body.
func NewBlob(body []byte) *Blob {
	return &Blob{
		body: slices.Clone(body),
	}
}

// Type returns ObjectTypeBlob.
func (b *Blob) Type() ObjectType {
	return ObjectTypeBlob
}

// Size returns the size of the blob contents in bytes.
func (b *Blob) Size() int {
	return len(b.body)
}

// Body returns a defensive copy of the blob contents.
func (b *Blob) Body() []byte {
	return slices.Clone(b.body)
}
