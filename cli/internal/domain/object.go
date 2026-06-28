package domain

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

var (
	// ErrInvalidObject indicates malformed or inconsistent serialized object data.
	ErrInvalidObject = errors.New("invalid object")
)

// Object represents an object stored in the Gel object database.
type Object interface {
	// Type returns the object's type.
	Type() ObjectType

	// Size returns the byte length of the serialized body.
	Size() int

	// Serialize returns the object in "<type> <size>\x00<body>" format.
	Serialize() []byte

	// Body returns a defensive copy of the serialized body.
	Body() []byte
}

// DeserializeObject parses data in "<type> <size>\x00<body>" format and
// constructs the corresponding object. It returns an error matching
// ErrInvalidObject when data is malformed or inconsistent.
func DeserializeObject(data []byte) (Object, error) {
	terminatorIndex := bytes.IndexByte(data, 0)
	if terminatorIndex == -1 {
		return nil, fmt.Errorf(
			"%w: missing header terminator",
			ErrInvalidObject,
		)
	}

	objectType, size, err := parseObjectHeader(data[:terminatorIndex])
	if err != nil {
		return nil, err
	}

	body := data[terminatorIndex+1:]
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
		return ParseTree(body)
	case ObjectTypeCommit:
		return NewCommit(body)
	default:
		return nil, fmt.Errorf(
			"%w: unknown object type %q",
			ErrInvalidObject,
			objectType,
		)
	}
}

func serializeObject(objectType ObjectType, body []byte) []byte {
	var buf bytes.Buffer
	buf.WriteString(objectType.String())
	buf.WriteByte(' ')
	buf.WriteString(strconv.Itoa(len(body)))
	buf.WriteByte(0)
	buf.Write(body)
	return buf.Bytes()
}

// parseObjectHeader parses a "<type> <size>" object header.
func parseObjectHeader(data []byte) (ObjectType, int, error) {
	spaceIndex := bytes.IndexByte(data, ' ')
	if spaceIndex == -1 {
		return "", 0, fmt.Errorf(
			"%w: missing type/size separator",
			ErrInvalidObject,
		)
	}

	objectTypeName := string(data[:spaceIndex])
	objectType, ok := ParseObjectType(objectTypeName)
	if !ok {
		return "", 0, fmt.Errorf(
			"%w: unknown object type %q",
			ErrInvalidObject,
			objectTypeName,
		)
	}

	size, err := parseObjectSize(data[spaceIndex+1:])
	if err != nil {
		return "", 0, err
	}
	return objectType, size, nil
}

// parseObjectSize parses a non-negative decimal object size.
func parseObjectSize(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("%w: empty object size", ErrInvalidObject)
	}

	for _, b := range data {
		if b < '0' || b > '9' {
			return 0, fmt.Errorf(
				"%w: invalid object size %q",
				ErrInvalidObject,
				data,
			)
		}
	}

	size, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, fmt.Errorf(
			"%w: object size %q is out of range",
			ErrInvalidObject,
			data,
		)
	}
	return size, nil
}
