package nbt

import (
	"errors"
	"fmt"
	"math"
	"reflect"
)

var tagInterface = reflect.TypeFor[Tag]()

var errRootNotCompound = errors.New("nbt: root tag is not a compound")

func Unmarshal(data []byte, v any) error {
	_, err := unmarshalRoot(data, false, v)

	return err
}

func UnmarshalNamed(data []byte, v any) (string, error) {
	return unmarshalRoot(data, true, v)
}

func unmarshalRoot(data []byte, named bool, v any) (string, error) {
	target := reflect.ValueOf(v)
	if target.Kind() != reflect.Pointer || target.IsNil() {
		return "", fmt.Errorf("nbt: unmarshal target must be a non-nil pointer, got %T", v)
	}

	dec := &decoder{
		buf: data,
	}
	if TagType(dec.u8()) != TagCompound {
		if dec.err != nil {
			return "", dec.err
		}
		return "", errRootNotCompound
	}

	name := ""
	if named {
		name = dec.str()
	}

	u := unmarshaler{
		dec: dec,
	}
	u.compound(target.Elem())

	return name, dec.err
}

type unmarshaler struct {
	dec *decoder
}

func (u *unmarshaler) compound(target reflect.Value) {
	if !u.dec.enter() {
		return
	}
	defer u.dec.leave()

	target = settable(target)

	switch target.Kind() {
	case reflect.Struct:
		u.readStruct(target)
	case reflect.Map:
		u.readMap(target)
	case reflect.Interface:
		if tree := u.dec.compound(); u.dec.err == nil {
			target.Set(reflect.ValueOf(tree))
		}
	default:
		u.dec.fail(fmt.Errorf("nbt: cannot unmarshal compound into %s", target.Kind()))
	}
}

func (u *unmarshaler) readStruct(target reflect.Value) {
	fields := fieldMapOf(target.Type())

	for {
		tag := TagType(u.dec.u8())
		if u.dec.err != nil || tag == TagEnd {
			return
		}

		name := u.dec.str()
		if f, ok := fields[name]; ok {
			u.value(tag, fieldTarget(target, f.index))
		} else {
			u.dec.payload(tag)
		}
	}
}

func (u *unmarshaler) readMap(target reflect.Value) {
	if target.Type().Key().Kind() != reflect.String {
		u.dec.fail(fmt.Errorf("nbt: map key must be string, got %s", target.Type().Key()))
		return
	}
	if target.IsNil() {
		target.Set(reflect.MakeMap(target.Type()))
	}

	elemType := target.Type().Elem()

	for {
		tag := TagType(u.dec.u8())
		if u.dec.err != nil || tag == TagEnd {
			return
		}

		name := u.dec.str()
		elem := reflect.New(elemType).Elem()
		u.value(tag, elem)
		if u.dec.err != nil {
			return
		}

		target.SetMapIndex(reflect.ValueOf(name), elem)
	}
}

func (u *unmarshaler) value(tag TagType, target reflect.Value) {
	if u.dec.err != nil {
		return
	}

	if target.Kind() == reflect.Pointer {
		if target.IsNil() {
			target.Set(reflect.New(target.Type().Elem()))
		}
		u.value(tag, target.Elem())
		return
	}

	if target.Kind() == reflect.Interface || target.Type().Implements(tagInterface) {
		u.dynamic(tag, target)
		return
	}

	switch tag {
	case TagByte:
		u.setInt(target, int64(int8(u.dec.u8())))
	case TagShort:
		u.setInt(target, int64(int16(u.dec.u16())))
	case TagInt:
		u.setInt(target, int64(int32(u.dec.u32())))
	case TagLong:
		u.setInt(target, int64(u.dec.u64()))
	case TagFloat:
		u.setFloat(target, float64(math.Float32frombits(u.dec.u32())))
	case TagDouble:
		u.setFloat(target, math.Float64frombits(u.dec.u64()))
	case TagString:
		u.setString(target, u.dec.str())
	case TagByteArray, TagIntArray, TagLongArray:
		u.array(tag, target)
	case TagList:
		u.list(target)
	case TagCompound:
		u.compound(target)
	default:
		u.dec.fail(fmt.Errorf("nbt: unknown tag %d", tag))
	}
}

func (u *unmarshaler) dynamic(tag TagType, target reflect.Value) {
	tree := u.dec.payload(tag)
	if u.dec.err != nil {
		return
	}

	value := reflect.ValueOf(tree)
	switch {
	case value.Type().AssignableTo(target.Type()):
		target.Set(value)
	case value.Type().ConvertibleTo(target.Type()):
		target.Set(value.Convert(target.Type()))
	default:
		u.dec.fail(fmt.Errorf("nbt: cannot assign %s to %s", value.Type(), target.Type()))
	}
}

func (u *unmarshaler) list(target reflect.Value) {
	elem := TagType(u.dec.u8())
	n := u.dec.length()
	if u.dec.err != nil {
		return
	}

	if target.Kind() != reflect.Slice {
		for range n {
			u.dec.payload(elem)
		}
		u.dec.fail(fmt.Errorf("nbt: cannot unmarshal list into %s", target.Type()))
		return
	}

	slice := reflect.MakeSlice(target.Type(), 0, min(n, maxPrealloc))
	for range n {
		item := reflect.New(target.Type().Elem()).Elem()
		u.value(elem, item)
		if u.dec.err != nil {
			return
		}
		slice = reflect.Append(slice, item)
	}

	target.Set(slice)
}

func (u *unmarshaler) array(tag TagType, target reflect.Value) {
	n := u.dec.length()
	if u.dec.err != nil {
		return
	}

	if target.Kind() != reflect.Slice {
		u.skipArray(tag, n)
		u.dec.fail(fmt.Errorf("nbt: cannot unmarshal array into %s", target.Type()))
		return
	}

	elem := target.Type().Elem()
	slice := reflect.MakeSlice(target.Type(), 0, min(n, maxPrealloc))
	for range n {
		item := reflect.New(elem).Elem()
		switch tag {
		case TagByteArray:
			item.SetInt(int64(int8(u.dec.u8())))
		case TagIntArray:
			item.SetInt(int64(int32(u.dec.u32())))
		case TagLongArray:
			item.SetInt(int64(u.dec.u64()))
		}
		if u.dec.err != nil {
			return
		}
		slice = reflect.Append(slice, item)
	}

	target.Set(slice)
}

var arrayElementWidth = map[TagType]int{
	TagByteArray: 1,
	TagIntArray:  4,
	TagLongArray: 8,
}

func (u *unmarshaler) skipArray(tag TagType, n int) {
	u.dec.take(n * arrayElementWidth[tag])
}

func (u *unmarshaler) setInt(target reflect.Value, value int64) {
	switch target.Kind() {
	case reflect.Bool:
		target.SetBool(value != 0)
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		target.SetInt(value)
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		target.SetUint(uint64(value))
	default:
		u.dec.fail(fmt.Errorf("nbt: cannot assign integer to %s", target.Type()))
	}
}

func (u *unmarshaler) setFloat(target reflect.Value, value float64) {
	if target.Kind() == reflect.Float32 || target.Kind() == reflect.Float64 {
		target.SetFloat(value)
		return
	}

	u.dec.fail(fmt.Errorf("nbt: cannot assign float to %s", target.Type()))
}

func (u *unmarshaler) setString(target reflect.Value, value string) {
	if target.Kind() == reflect.String {
		target.SetString(value)
		return
	}

	u.dec.fail(fmt.Errorf("nbt: cannot assign string to %s", target.Type()))
}

func settable(target reflect.Value) reflect.Value {
	for target.Kind() == reflect.Pointer {
		if target.IsNil() {
			target.Set(reflect.New(target.Type().Elem()))
		}
		target = target.Elem()
	}

	return target
}
