package graft_test

import (
	"math"
	"testing"

	graft "github.com/lrnxzz/graft"
)

type lookCase struct {
	yaw   float32
	pitch float32
	want  graft.Vec3d
}

func TestPlayerLookDirection(t *testing.T) {
	cases := []lookCase{
		{
			yaw:   0,
			pitch: 0,
			want:  graft.Vec3(0, 0, 1),
		},
		{
			yaw:   90,
			pitch: 0,
			want:  graft.Vec3(-1, 0, 0),
		},
		{
			yaw:   -90,
			pitch: 0,
			want:  graft.Vec3(1, 0, 0),
		},
		{
			yaw:   0,
			pitch: 90,
			want:  graft.Vec3(0, -1, 0),
		},
		{
			yaw:   0,
			pitch: -90,
			want:  graft.Vec3(0, 1, 0),
		},
	}
	for _, c := range cases {
		player := &graft.Player{
			Yaw:   c.yaw,
			Pitch: c.pitch,
		}

		got := player.LookDirection()
		if !got.ApproxEqual(c.want, 1e-9) {
			t.Errorf("LookDirection(yaw %v, pitch %v) = %v, want %v", c.yaw, c.pitch, got, c.want)
		}
	}
}

func TestPlayerEye(t *testing.T) {
	player := &graft.Player{
		Position: graft.Vec3(1, 64, -3),
	}

	got := player.Eye()
	want := graft.Vec3(1, 65.62, -3)
	if !got.ApproxEqual(want, 1e-9) {
		t.Errorf("Eye() = %v, want %v", got, want)
	}
}

func TestPlayerBox(t *testing.T) {
	player := &graft.Player{
		Position: graft.Vec3(0, 10, 0),
	}

	box := player.Box()
	if box.Min.Y != 10 || box.Max.Y != 11.8 {
		t.Errorf("box spans Y [%v, %v], want [10, 11.8]", box.Min.Y, box.Max.Y)
	}
	if width := box.Max.X - box.Min.X; math.Abs(width-0.6) > 1e-9 {
		t.Errorf("box width = %v, want 0.6", width)
	}
}

func TestPlayerAlive(t *testing.T) {
	player := &graft.Player{
		Health: 20,
	}
	if !player.Alive() {
		t.Error("player with full health should be alive")
	}

	player.Health = 0
	if player.Alive() {
		t.Error("player with zero health should be dead")
	}
}

func TestLookAnglesMatchLookDirection(t *testing.T) {
	from := graft.Vec3(0.5, 65, 0.5)
	targets := []graft.Vec3d{
		graft.Vec3(10, 65, 0.5),
		graft.Vec3(0.5, 70, 10),
		graft.Vec3(-3, 60, -8),
		graft.Vec3(0.5, 80, 0.5),
	}
	for _, target := range targets {
		yaw, pitch := graft.LookAngles(from, target)
		player := &graft.Player{
			Yaw:   yaw,
			Pitch: pitch,
		}

		got := player.LookDirection()
		want := target.Sub(from).Normalize()
		if !got.ApproxEqual(want, 1e-6) {
			t.Errorf("LookAngles(%v) round trip = %v, want %v", target, got, want)
		}
	}
}
