package graft_test

import (
	"math"
	"testing"

	graft "github.com/lrnxzz/graft"
)

// zero yaw looks toward +Z and the angle grows clockwise, which is the protocol's
// convention and regresses silently if a sign flips
func TestYawFollowsTheProtocolConvention(t *testing.T) {
	tests := []struct {
		heading graft.Vec2d
		yaw     float32
	}{
		{
			heading: graft.Vec2(0, 1),
			yaw:     0,
		},
		{
			heading: graft.Vec2(-1, 0),
			yaw:     90,
		},
		{
			heading: graft.Vec2(1, 0),
			yaw:     -90,
		},
		{
			// atan2 sees the negative zero in -X, so straight back is -180
			heading: graft.Vec2(0, -1),
			yaw:     -180,
		},
	}

	for _, tt := range tests {
		if got := tt.heading.Yaw(); got != tt.yaw {
			t.Errorf("heading %s yields yaw %g, want %g", tt.heading, got, tt.yaw)
		}
	}
}

// a negative coordinate floors away from zero: the block at -0.5 is block -1
func TestFloorOnNegativeCoordinates(t *testing.T) {
	tests := []struct {
		at   graft.Vec3d
		want graft.Position
	}{
		{
			at:   graft.Vec3(0.5, 64.9, 0.1),
			want: graft.At(0, 64, 0),
		},
		{
			at:   graft.Vec3(-0.5, -0.1, -1.9),
			want: graft.At(-1, -1, -2),
		},
		{
			at:   graft.Vec3(-1, 0, 1),
			want: graft.At(-1, 0, 1),
		},
	}

	for _, tt := range tests {
		if got := tt.at.Floor(); got != tt.want {
			t.Errorf("%s floors to %s, want %s", tt.at, got, tt.want)
		}
	}
}

func TestNormalize(t *testing.T) {
	unit := graft.Vec3(3, 0, 4).Normalize()
	if !unit.ApproxEqual(graft.Vec3(0.6, 0, 0.8), 1e-12) {
		t.Errorf("normalized to %s, want (0.6, 0, 0.8)", unit)
	}

	if length := unit.Length(); math.Abs(length-1) > 1e-12 {
		t.Errorf("normalized length = %g, want 1", length)
	}

	if zero := (graft.Vec3d{}).Normalize(); zero != (graft.Vec3d{}) {
		t.Errorf("the zero vector normalized to %s, want itself", zero)
	}
}
