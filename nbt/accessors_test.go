package nbt_test

import (
	"reflect"
	"testing"

	"github.com/lrnxzz/graft/nbt"
)

func TestGet(t *testing.T) {
	compound := nbt.Compound{
		"height": nbt.Int(320),
		"name":   nbt.String("overworld"),
	}

	height, ok := nbt.Get[nbt.Int](compound, "height")
	if !ok || height != 320 {
		t.Errorf("Get[Int](height) = (%d, %t), want (320, true)", height, ok)
	}

	if _, ok := nbt.Get[nbt.String](compound, "height"); ok {
		t.Error("Get[String](height) matched an Int value")
	}

	if _, ok := nbt.Get[nbt.Int](compound, "missing"); ok {
		t.Error("Get[Int](missing) matched an absent key")
	}
}

func TestItems(t *testing.T) {
	homogeneous := nbt.List{
		Elem:  nbt.TagDouble,
		Items: []nbt.Tag{nbt.Double(1.5), nbt.Double(2.5), nbt.Double(3.5)},
	}

	doubles, ok := nbt.Items[nbt.Double](homogeneous)
	if !ok {
		t.Fatal("Items[Double] failed on a homogeneous double list")
	}

	want := []nbt.Double{1.5, 2.5, 3.5}
	if !reflect.DeepEqual(doubles, want) {
		t.Errorf("Items[Double] = %v, want %v", doubles, want)
	}

	mixed := nbt.List{
		Elem:  nbt.TagInt,
		Items: []nbt.Tag{nbt.Int(1), nbt.String("x")},
	}

	if _, ok := nbt.Items[nbt.Int](mixed); ok {
		t.Error("Items[Int] accepted a list containing a String")
	}
}
