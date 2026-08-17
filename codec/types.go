package codec

import (
	"fmt"
	"math"
	"unsafe"
)

type unsigned interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64
}

const (
	MaxStringLen   = 32767
	maxStringBytes = 3*MaxStringLen + 3
	maxPrealloc    = 4096
)

type Field interface {
	Append(dst []byte) []byte
}

type FieldPtr interface {
	Decode(r *Reader) error
}

func AppendAll(dst []byte, fields ...Field) []byte {
	for _, field := range fields {
		dst = field.Append(dst)
	}

	return dst
}

func DecodeAll(r *Reader, fields ...FieldPtr) error {
	for _, field := range fields {
		if err := field.Decode(r); err != nil {
			return err
		}
	}

	return nil
}

func Marshal(fields ...Field) []byte {
	return AppendAll(nil, fields...)
}

func Unmarshal(payload []byte, fields ...FieldPtr) error {
	return DecodeAll(NewReader(payload), fields...)
}

func appendBE[T unsigned](dst []byte, v T) []byte {
	for shift := (int(unsafe.Sizeof(v)) - 1) * bitsPerByte; shift >= 0; shift -= bitsPerByte {
		dst = append(dst, byte(v>>shift))
	}
	return dst
}

func readBE[T unsigned](r *Reader) (T, error) {
	raw := r.take(int(unsafe.Sizeof(T(0))))
	if raw == nil {
		return 0, r.err
	}
	var v T
	for i, octet := range raw {
		v |= T(octet) << ((len(raw) - 1 - i) * bitsPerByte)
	}
	return v, nil
}

type (
	Bool    bool
	Byte    int8
	UByte   uint8
	Short   int16
	UShort  uint16
	Int     int32
	Long    int64
	Float   float32
	Double  float64
	VarInt  int32
	VarLong int64
	String  string
	UUID    [16]byte
)

func (v Bool) Append(dst []byte) []byte {
	var raw uint8
	if v {
		raw = 1
	}
	return appendBE(dst, raw)
}

func (v *Bool) Decode(r *Reader) error {
	raw, err := readBE[uint8](r)
	if err != nil {
		return err
	}
	*v = raw != 0
	return nil
}

func (v Bool) Bool() bool {
	return bool(v)
}

func (v Byte) Append(dst []byte) []byte {
	return appendBE(dst, uint8(v))
}

func (v *Byte) Decode(r *Reader) error {
	raw, err := readBE[uint8](r)
	if err != nil {
		return err
	}
	*v = Byte(raw)
	return nil
}

func (v Byte) Int8() int8 {
	return int8(v)
}

func (v Byte) Int() int {
	return int(v)
}

func (v UByte) Append(dst []byte) []byte {
	return appendBE(dst, uint8(v))
}

func (v *UByte) Decode(r *Reader) error {
	raw, err := readBE[uint8](r)
	if err != nil {
		return err
	}
	*v = UByte(raw)
	return nil
}

func (v UByte) Uint8() uint8 {
	return uint8(v)
}

func (v UByte) Int() int {
	return int(v)
}

func (v Short) Append(dst []byte) []byte {
	return appendBE(dst, uint16(v))
}

func (v *Short) Decode(r *Reader) error {
	raw, err := readBE[uint16](r)
	if err != nil {
		return err
	}
	*v = Short(raw)
	return nil
}

func (v Short) Int16() int16 {
	return int16(v)
}

func (v Short) Int() int {
	return int(v)
}

func (v UShort) Append(dst []byte) []byte {
	return appendBE(dst, uint16(v))
}

func (v *UShort) Decode(r *Reader) error {
	raw, err := readBE[uint16](r)
	if err != nil {
		return err
	}
	*v = UShort(raw)
	return nil
}

func (v UShort) Uint16() uint16 {
	return uint16(v)
}

func (v UShort) Int() int {
	return int(v)
}

func (v Int) Append(dst []byte) []byte {
	return appendBE(dst, uint32(v))
}

func (v *Int) Decode(r *Reader) error {
	raw, err := readBE[uint32](r)
	if err != nil {
		return err
	}
	*v = Int(raw)
	return nil
}

func (v Int) Int32() int32 {
	return int32(v)
}

func (v Int) Int64() int64 {
	return int64(v)
}

func (v Int) Int() int {
	return int(v)
}

func (v Long) Append(dst []byte) []byte {
	return appendBE(dst, uint64(v))
}

func (v *Long) Decode(r *Reader) error {
	raw, err := readBE[uint64](r)
	if err != nil {
		return err
	}
	*v = Long(raw)
	return nil
}

func (v Long) Int64() int64 {
	return int64(v)
}

func (v Long) Int() int {
	return int(v)
}

func (v Long) Uint64() uint64 {
	return uint64(v)
}

func (v Long) Signed(offset, size int) int64 {
	return int64(v) << (64 - offset - size) >> (64 - size)
}

func (v Long) Unsigned(offset, size int) int64 {
	return int64(v) >> offset & (int64(1)<<size - 1)
}

func (v Float) Append(dst []byte) []byte {
	return appendBE(dst, math.Float32bits(float32(v)))
}

func (v *Float) Decode(r *Reader) error {
	raw, err := readBE[uint32](r)
	if err != nil {
		return err
	}
	*v = Float(math.Float32frombits(raw))
	return nil
}

func (v Float) Float32() float32 {
	return float32(v)
}

func (v Float) Float64() float64 {
	return float64(v)
}

func (v Double) Append(dst []byte) []byte {
	return appendBE(dst, math.Float64bits(float64(v)))
}

func (v *Double) Decode(r *Reader) error {
	raw, err := readBE[uint64](r)
	if err != nil {
		return err
	}
	*v = Double(math.Float64frombits(raw))
	return nil
}

func (v Double) Float64() float64 {
	return float64(v)
}

func (v VarInt) Append(dst []byte) []byte {
	return AppendVar(dst, v)
}

func (v *VarInt) Decode(r *Reader) error {
	val, err := ReadVar[VarInt](r)
	if err != nil {
		return r.fail(err)
	}
	*v = val
	return nil
}

func (v VarInt) Int32() int32 {
	return int32(v)
}

func (v VarInt) Int64() int64 {
	return int64(v)
}

func (v VarInt) Int() int {
	return int(v)
}

func (v VarLong) Append(dst []byte) []byte {
	return AppendVar(dst, v)
}

func (v *VarLong) Decode(r *Reader) error {
	val, err := ReadVar[VarLong](r)
	if err != nil {
		return r.fail(err)
	}
	*v = val
	return nil
}

func (v VarLong) Int64() int64 {
	return int64(v)
}

func (v VarLong) Int() int {
	return int(v)
}

func (v String) Append(dst []byte) []byte {
	dst = AppendVar(dst, VarInt(len(v)))
	return append(dst, v...)
}

func (v *String) Decode(r *Reader) error {
	n, err := ReadVar[VarInt](r)
	if err != nil {
		return r.fail(err)
	}
	if n < 0 || n > maxStringBytes {
		return r.fail(fmt.Errorf("graft: string of %d bytes is out of range", n))
	}
	raw := r.take(int(n))
	if raw == nil {
		return r.err
	}
	*v = String(raw)
	return nil
}

func (v String) String() string {
	return string(v)
}

func (v UUID) Append(dst []byte) []byte {
	return append(dst, v[:]...)
}

func (v *UUID) Decode(r *Reader) error {
	raw := r.take(len(v))
	if raw == nil {
		return r.err
	}
	copy(v[:], raw)
	return nil
}

type Slice[T Field] []T

func (s Slice[T]) Append(dst []byte) []byte {
	dst = AppendVar(dst, VarInt(len(s)))
	for _, element := range s {
		dst = element.Append(dst)
	}
	return dst
}

func (s *Slice[T]) Decode(r *Reader) error {
	var n VarInt
	if err := n.Decode(r); err != nil {
		return err
	}
	if n < 0 || int(n) > r.Remaining() {
		return r.fail(fmt.Errorf("graft: slice of %d elements is out of range", n))
	}
	elements := make(Slice[T], 0, min(int(n), maxPrealloc))
	for range int(n) {
		var element T
		ptr, ok := any(&element).(FieldPtr)
		if !ok {
			return r.fail(fmt.Errorf("graft: %T does not implement FieldPtr", &element))
		}
		if err := ptr.Decode(r); err != nil {
			return err
		}
		elements = append(elements, element)
	}
	*s = elements
	return nil
}

type Bytes []byte

func (b Bytes) Append(dst []byte) []byte {
	dst = AppendVar(dst, VarInt(len(b)))
	return append(dst, b...)
}

func (b *Bytes) Decode(r *Reader) error {
	var n VarInt
	if err := n.Decode(r); err != nil {
		return err
	}
	raw := r.take(int(n))
	if raw == nil {
		return r.err
	}
	*b = append(Bytes(nil), raw...)
	return nil
}

type Option[T Field] struct {
	value   T
	present bool
}

func Some[T Field](v T) Option[T] {
	return Option[T]{
		value:   v,
		present: true,
	}
}

func None[T Field]() Option[T] {
	return Option[T]{}
}

func (o Option[T]) Get() (T, bool) {
	return o.value, o.present
}

func (o Option[T]) Append(dst []byte) []byte {
	dst = Bool(o.present).Append(dst)
	if o.present {
		dst = o.value.Append(dst)
	}
	return dst
}

func (o *Option[T]) Decode(r *Reader) error {
	var present Bool
	if err := present.Decode(r); err != nil {
		return err
	}
	o.present = bool(present)
	if !o.present {
		var zero T
		o.value = zero
		return nil
	}
	ptr, ok := any(&o.value).(FieldPtr)
	if !ok {
		return r.fail(fmt.Errorf("graft: %T does not implement FieldPtr", &o.value))
	}
	return ptr.Decode(r)
}
