package nbt

import (
	"reflect"
	"slices"
	"strings"
	"sync"
)

type field struct {
	name      string
	index     []int
	omitempty bool
	asList    bool
}

var (
	_fieldCache    sync.Map
	_fieldMapCache sync.Map
)

func fieldsOf(t reflect.Type) []field {
	cached, hit := _fieldCache.Load(t)
	if hit {
		fields, ok := cached.([]field)
		if ok {
			return fields
		}
	}

	fields := resolveShadows(collectFields(t, nil))
	_fieldCache.Store(t, fields)

	return fields
}

func fieldMapOf(t reflect.Type) map[string]field {
	cached, hit := _fieldMapCache.Load(t)
	if hit {
		byName, ok := cached.(map[string]field)
		if ok {
			return byName
		}
	}

	byName := make(map[string]field, t.NumField())
	for _, f := range fieldsOf(t) {
		byName[f.name] = f
	}
	_fieldMapCache.Store(t, byName)

	return byName
}

func collectFields(t reflect.Type, prefix []int) []field {
	var fields []field

	for i := range t.NumField() {
		structField := t.Field(i)

		tag, tagged := structField.Tag.Lookup("nbt")
		if tag == "-" {
			continue
		}

		promoted := structField.Anonymous && !tagged && structType(structField.Type).Kind() == reflect.Struct
		if !structField.IsExported() && !promoted {
			continue
		}

		index := append(slices.Clone(prefix), i)

		if promoted {
			fields = append(fields, collectFields(structType(structField.Type), index)...)
			continue
		}

		name, opts, _ := strings.Cut(tag, ",")
		if name == "" {
			name = structField.Name
		}

		options := strings.Split(opts, ",")
		fields = append(fields, field{
			name:      name,
			index:     index,
			omitempty: slices.Contains(options, "omitempty"),
			asList:    slices.Contains(options, "list"),
		})
	}

	return fields
}

type shallowest struct {
	depth int
	count int
}

func resolveShadows(fields []field) []field {
	winners := make(map[string]shallowest, len(fields))
	for _, f := range fields {
		depth := len(f.index)
		current, ok := winners[f.name]
		switch {
		case !ok || depth < current.depth:
			winners[f.name] = shallowest{depth: depth, count: 1}
		case depth == current.depth:
			current.count++
			winners[f.name] = current
		}
	}

	resolved := fields[:0]
	for _, f := range fields {
		winner := winners[f.name]
		if len(f.index) == winner.depth && winner.count == 1 {
			resolved = append(resolved, f)
		}
	}

	return resolved
}

func structType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Pointer {
		return t.Elem()
	}

	return t
}

func fieldValue(v reflect.Value, index []int) reflect.Value {
	for _, i := range index {
		if v.Kind() == reflect.Pointer {
			if v.IsNil() {
				return reflect.Value{}
			}
			v = v.Elem()
		}
		v = v.Field(i)
	}

	return v
}

func fieldTarget(v reflect.Value, index []int) reflect.Value {
	for _, i := range index {
		if v.Kind() == reflect.Pointer {
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
		}
		v = v.Field(i)
	}

	return v
}

func deref(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return v
		}
		v = v.Elem()
	}

	return v
}
