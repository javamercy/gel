package domain

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

// ErrInvalidObject indicates malformed or inconsistent serialized object data.
var ErrInvalidObject = errors.New("invalid gel object")

// Object represents an object stored in the Gel object database.
type Object interface {
	// Type returns the object's type.
	Type() ObjectType

	// Size returns the byte length of the serialized body.
	Size() int

	// Body returns a defensive copy of the serialized body.
	Body() []byte
}

// EncodeObject encodes an object in "<type> <size>\x00<body>" format.
func EncodeObject(object Object) ([]byte, error) {
	if object == nil {
		return nil, fmt.Errorf(
			"%w: object is nil",
			ErrInvalidObject,
		)
	}

	objectType := object.Type()
	if !objectType.IsValid() {
		return nil, fmt.Errorf(
			"%w: unsupported object type %q",
			ErrInvalidObject,
			objectType,
		)
	}

	body := object.Body()
	if object.Size() != len(body) {
		return nil, fmt.Errorf(
			"%w: body size mismatch: reported=%d actual=%d",
			ErrInvalidObject,
			object.Size(),
			len(body),
		)
	}

	sizeText := strconv.Itoa(len(body))
	encoded := make([]byte, 0, len(objectType.String())+1+len(sizeText)+1+len(body))

	encoded = append(encoded, objectType.String()...)
	encoded = append(encoded, ' ')
	encoded = append(encoded, sizeText...)
	encoded = append(encoded, 0)
	encoded = append(encoded, body...)
	return encoded, nil
}

// DecodeObject decodes data in "<type> <size>\x00<body>" format.
//
// It returns an error matching ErrInvalidObject when data is malformed or
// inconsistent.
func DecodeObject(data []byte) (Object, error) {
	nulIndex := bytes.IndexByte(data, 0)
	if nulIndex == -1 {
		return nil, fmt.Errorf(
			"%w: missing header terminator",
			ErrInvalidObject,
		)
	}

	objectType, size, err := parseObjectHeader(data[:nulIndex])
	if err != nil {
		return nil, fmt.Errorf(
			"%w: decode header: %w",
			ErrInvalidObject,
			err,
		)
	}

	body := data[nulIndex+1:]
	if len(body) != size {
		return nil, fmt.Errorf(
			"%w: body size mismatch: declared=%d actual=%d",
			ErrInvalidObject,
			size,
			len(body),
		)
	}

	switch objectType {
	case ObjectTypeBlob:
		return NewBlob(body), nil
	case ObjectTypeTree:
		tree, err := DecodeTree(body)
		if err != nil {
			return nil, fmt.Errorf(
				"%w: decode tree body: %w",
				ErrInvalidObject,
				err,
			)
		}
		return tree, nil
	case ObjectTypeCommit:
		commit, err := NewCommit(body)
		if err != nil {
			return nil, fmt.Errorf(
				"%w: decode commit body: %w",
				ErrInvalidObject,
				err,
			)
		}
		return commit, nil
	default:
		return nil, fmt.Errorf(
			"%w: unsupported object type %q",
			ErrInvalidObject,
			objectType,
		)
	}
}

func parseObjectHeader(data []byte) (objectType ObjectType, size int, err error) {
	spaceIndex := bytes.IndexByte(data, ' ')
	if spaceIndex == -1 {
		return "", 0, fmt.Errorf("missing type/size separator")
	}

	objectTypeName := string(data[:spaceIndex])
	objectType, ok := ParseObjectType(objectTypeName)
	if !ok {
		return "", 0, fmt.Errorf("unknown object type %q", objectTypeName)
	}

	size, err = parseObjectSize(data[spaceIndex+1:])
	if err != nil {
		return "", 0, err
	}
	return objectType, size, nil
}

func parseObjectSize(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("empty object size")
	}
	for _, b := range data {
		if b < '0' || b > '9' {
			return 0, fmt.Errorf("invalid object size %q", data)
		}
	}

	size, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, fmt.Errorf("object size %q is out of range", data)
	}
	return size, nil
}
