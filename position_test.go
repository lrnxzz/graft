package graft_test

import (
	"testing"

	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/codec"
)

type blockFaceCase struct {
	face graft.BlockFace
	want graft.Position
}

func TestBlockFaceNeighbors(t *testing.T) {
	origin := graft.Position{
		X: 10,
		Y: 64,
		Z: -3,
	}

	cases := []blockFaceCase{
		{
			face: graft.FaceUp,
			want: origin.Add(0, 1, 0),
		},
		{
			face: graft.FaceDown,
			want: origin.Add(0, -1, 0),
		},
		{
			face: graft.FaceNorth,
			want: origin.Add(0, 0, -1),
		},
		{
			face: graft.FaceSouth,
			want: origin.Add(0, 0, 1),
		},
		{
			face: graft.FaceWest,
			want: origin.Add(-1, 0, 0),
		},
		{
			face: graft.FaceEast,
			want: origin.Add(1, 0, 0),
		},
	}
	for _, c := range cases {
		got := origin.Neighbor(c.face)
		if got != c.want {
			t.Errorf("Neighbor(%v) = %v, want %v", c.face, got, c.want)
		}

		back := got.Neighbor(c.face.Opposite())
		if back != origin {
			t.Errorf("round trip through %v and %v = %v, want %v", c.face, c.face.Opposite(), back, origin)
		}
	}
}

func TestPositionCenter(t *testing.T) {
	block := graft.Position{
		X: 1,
		Y: 2,
		Z: -4,
	}

	got := block.Center()
	want := graft.Vec3(1.5, 2.5, -3.5)
	if got != want {
		t.Errorf("Center() = %v, want %v", got, want)
	}
}

func TestPositionRecoversPackedCoordinates(t *testing.T) {
	positions := []graft.Position{
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 2, Z: 3},
		{X: -1, Y: -1, Z: -1},
		{X: 33554431, Y: 2047, Z: 33554431},
		{X: -33554432, Y: -2048, Z: -33554432},
		{X: 33554431, Y: -2048, Z: -33554432},
	}

	for _, want := range positions {
		var got graft.Position

		if err := codec.Unmarshal(want.Append(nil), &got); err != nil {
			t.Errorf("decode %s: %v", want, err)
			continue
		}
		if got != want {
			t.Errorf("round trip of %s yielded %s", want, got)
		}
	}
}

func TestAngleDegrees(t *testing.T) {
	tests := []struct {
		angle   graft.Angle
		degrees float64
	}{
		{
			angle:   0,
			degrees: 0,
		},
		{
			angle:   64,
			degrees: 90,
		},
		{
			angle:   128,
			degrees: 180,
		},
		{
			angle:   192,
			degrees: 270,
		},
	}

	for _, tt := range tests {
		if got := tt.angle.Degrees(); got != tt.degrees {
			t.Errorf("Angle(%d).Degrees() = %g, want %g", tt.angle, got, tt.degrees)
		}
		if got := graft.AngleFromDegrees(tt.degrees); got != tt.angle {
			t.Errorf("AngleFromDegrees(%g) = %d, want %d", tt.degrees, got, tt.angle)
		}
	}
}

func TestAngleRecoversEncodedValue(t *testing.T) {
	for _, want := range []graft.Angle{0, 1, 64, 128, 200, 255} {
		var got graft.Angle

		if err := codec.Unmarshal(want.Append(nil), &got); err != nil {
			t.Errorf("decode %d: %v", want, err)
			continue
		}
		if got != want {
			t.Errorf("round trip of %d yielded %d", want, got)
		}
	}
}
