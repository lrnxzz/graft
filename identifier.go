package gocraft

import (
	"strings"

	"github.com/lrnxzz/go-craft/codec"
)

const DefaultNamespace = "minecraft"

type Identifier string

func NewIdentifier(namespace, path string) Identifier {
	return Identifier(namespace + ":" + path)
}

func (i Identifier) Append(dst []byte) []byte {
	return codec.String(i).Append(dst)
}

func (i *Identifier) Decode(r *codec.Reader) error {
	var raw codec.String
	if err := raw.Decode(r); err != nil {
		return err
	}

	*i = Identifier(raw)

	return nil
}

func (i Identifier) Namespace() string {
	if namespace, _, ok := strings.Cut(string(i), ":"); ok {
		return namespace
	}

	return DefaultNamespace
}

func (i Identifier) Path() string {
	if _, path, ok := strings.Cut(string(i), ":"); ok {
		return path
	}

	return string(i)
}

func (i Identifier) Valid() bool {
	namespace, path := i.Namespace(), i.Path()
	if path == "" {
		return false
	}

	return validSegment(namespace, validNamespaceByte) && validSegment(path, validPathByte)
}

func (i Identifier) String() string {
	return i.Namespace() + ":" + i.Path()
}

func validSegment(segment string, allow func(byte) bool) bool {
	if segment == "" {
		return false
	}

	for i := range len(segment) {
		if !allow(segment[i]) {
			return false
		}
	}

	return true
}

func validNamespaceByte(b byte) bool {
	return b >= 'a' && b <= 'z' || b >= '0' && b <= '9' || b == '_' || b == '-' || b == '.'
}

func validPathByte(b byte) bool {
	return validNamespaceByte(b) || b == '/'
}
