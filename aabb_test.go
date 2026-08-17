package graft_test

import (
	"testing"

	graft "github.com/lrnxzz/graft"
)

func TestAABBIntersects(t *testing.T) {
	block := graft.Box(graft.Vec3(0, 0, 0), graft.Vec3(1, 1, 1))
	overlapping := graft.Box(graft.Vec3(0.5, 0.5, 0.5), graft.Vec3(1.5, 1.5, 1.5))
	touching := graft.Box(graft.Vec3(1, 0, 0), graft.Vec3(2, 1, 1))

	if !block.Intersects(overlapping) {
		t.Error("overlapping boxes should intersect")
	}
	if block.Intersects(touching) {
		t.Error("boxes touching only on a face should not intersect")
	}
}

func TestAABBClampYStopsFall(t *testing.T) {
	ground := graft.Box(graft.Vec3(0, 0, 0), graft.Vec3(1, 1, 1))
	player := graft.BoxAround(graft.Vec3(0.5, 1.5, 0.5), 0.6, 1.8)

	if got := ground.ClampY(player, -1); got != -0.5 {
		t.Errorf("ClampY = %v, want -0.5", got)
	}
}

func TestAABBStretchSweepsAlongSign(t *testing.T) {
	swept := graft.Box(graft.Vec3(0, 0, 0), graft.Vec3(1, 1, 1)).Stretch(-2, 0, 3)

	if swept.Min.X != -2 || swept.Max.X != 1 {
		t.Errorf("X = [%v, %v], want [-2, 1]", swept.Min.X, swept.Max.X)
	}
	if swept.Min.Z != 0 || swept.Max.Z != 4 {
		t.Errorf("Z = [%v, %v], want [0, 4]", swept.Min.Z, swept.Max.Z)
	}
}
