package domain

// ObjectType identifies a domain object kind.
type ObjectType string

const (
	// ObjectTypeBlob represents a file content object.
	ObjectTypeBlob ObjectType = "blob"
	// ObjectTypeTree represents a directory structure object.
	ObjectTypeTree ObjectType = "tree"
	// ObjectTypeCommit represents a commit snapshot object.
	ObjectTypeCommit ObjectType = "commit"
)

// IsValid reports whether objectType is one of the supported values.
func (t ObjectType) IsValid() bool {
	switch t {
	case ObjectTypeBlob, ObjectTypeTree, ObjectTypeCommit:
		return true
	default:
		return false
	}
}

// String returns the underlying string value.
func (t ObjectType) String() string {
	return string(t)
}

// ParseObjectType converts typeStr to an ObjectType and reports whether it is valid.
func ParseObjectType(s string) (ObjectType, bool) {
	objectType := ObjectType(s)
	if objectType.IsValid() {
		return objectType, true
	}
	return "", false
}
