package domain

// Blob represents immutable file content stored as a blob object.
type Blob struct {
	body []byte
}

// NewBlob returns a Blob containing a defensive copy of body.
func NewBlob(body []byte) *Blob {
	return &Blob{
		body: append([]byte(nil), body...),
	}
}

// Body returns a defensive copy of the blob contents.
func (b *Blob) Body() []byte {
	return append([]byte(nil), b.body...)
}

// Type returns ObjectTypeBlob.
func (b *Blob) Type() ObjectType {
	return ObjectTypeBlob
}

// Size returns the size of the blob contents in bytes.
func (b *Blob) Size() int {
	return len(b.body)
}

// Serialize returns the full object serialization in the form "<type> <size>\x00<body>".
func (b *Blob) Serialize() []byte {
	return serializeObject(ObjectTypeBlob, b.body)
}
