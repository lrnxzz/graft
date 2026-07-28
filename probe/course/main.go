package main

import (
	"fmt"
	"log"
	"strings"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/rcon"
)

const (
	address  = "localhost:25575"
	password = "gocraft"

	// the flat preset caps at a grass block, so feet land one above it, and
	// nothing may be carved below the world floor or the server refuses the fill
	worldFloor = -64
	pitFloor   = -63
	ground     = -61
	stand      = -60

	laneMin = -5
	laneMax = 5
	fenceZ  = 6
	ceiling = -52

	courseStart = -6
	courseEnd   = 82
)

type builder struct {
	console *rcon.Client
	err     error
}

func (b *builder) run(command string) {
	if b.err != nil {
		return
	}

	answer, err := b.console.Run(command)
	if err != nil {
		b.err = err

		return
	}
	if rejected(answer) {
		b.err = fmt.Errorf("course: %q answered %q", command, answer)
	}
}

// the server answers a refused command with prose and no error, so a builder
// that does not read the reply happily reports a course it never carved
var refusals = [...]string{
	"Unknown",
	"Failed",
	"Incorrect",
	"Invalid",
	"Expected",
	"out of this world",
	"cannot be placed",
}

func rejected(answer string) bool {
	for _, refusal := range refusals {
		if strings.Contains(answer, refusal) {
			return true
		}
	}

	return false
}

func (b *builder) fill(from, to gocraft.Position, block string) {
	b.run(fmt.Sprintf("fill %d %d %d %d %d %d %s",
		from.X, from.Y, from.Z, to.X, to.Y, to.Z, block))
}

func (b *builder) slab(x1, x2, y int, block string) {
	b.fill(gocraft.At(x1, y, laneMin), gocraft.At(x2, y, laneMax), block)
}

func (b *builder) barrier(x1, x2, z int) {
	b.fill(gocraft.At(x1, stand, z), gocraft.At(x2, ceiling, z), "bedrock")
}

type station struct {
	name  string
	goal  gocraft.Position
	build func(*builder)
}

// every station is a walled corridor: without the fences the planner simply
// walks around the obstacle and never has to mine or bridge anything
var stations = [...]station{
	{
		name: "flat ground",
		goal: gocraft.At(6, stand, 0),
		build: func(*builder) {
		},
	},
	{
		name: "stairs up three",
		goal: gocraft.At(16, stand+3, 0),
		build: func(b *builder) {
			for step := range 3 {
				b.slab(11+step*2, 12+step*2, ground+1+step, "stone_bricks")
			}
			b.slab(17, 20, ground+3, "stone_bricks")
		},
	},
	{
		name: "three block drop",
		goal: gocraft.At(24, stand, 0),
		build: func(b *builder) {
			b.slab(21, 21, ground+3, "stone_bricks")
		},
	},
	{
		name: "parkour gap",
		goal: gocraft.At(34, stand, 0),
		build: func(b *builder) {
			b.fill(gocraft.At(29, pitFloor, laneMin), gocraft.At(31, ground, laneMax), "air")
		},
	},
	{
		name: "short dirt wall: climbs over, breaking the cap",
		goal: gocraft.At(42, stand, 0),
		build: func(b *builder) {
			b.fill(gocraft.At(38, stand, laneMin), gocraft.At(38, stand+1, laneMax), "dirt")
		},
	},
	{
		name: "tall dirt wall: tunnels through at foot level",
		goal: gocraft.At(50, stand, 0),
		build: func(b *builder) {
			b.fill(gocraft.At(46, stand, laneMin), gocraft.At(46, stand+5, laneMax), "dirt")
		},
	},
	{
		name: "stone wall with a gap: walking around beats mining",
		goal: gocraft.At(58, stand, 0),
		build: func(b *builder) {
			b.fill(gocraft.At(54, stand, laneMin), gocraft.At(54, stand+3, 3), "stone")
		},
	},
	{
		name: "four block chasm: leaps at the limit of the jump",
		goal: gocraft.At(70, stand, 0),
		build: func(b *builder) {
			b.fill(gocraft.At(63, pitFloor, laneMin), gocraft.At(66, ground, laneMax), "air")
		},
	},
	{
		name: "buried chamber: digs a shaft down",
		goal: gocraft.At(78, stand, 0),
		build: func(b *builder) {
			b.fill(gocraft.At(74, stand, laneMin), gocraft.At(80, stand+3, laneMax), "dirt")
			b.fill(gocraft.At(77, stand, -1), gocraft.At(79, stand+1, 1), "air")
			b.fill(gocraft.At(74, stand+4, laneMin), gocraft.At(80, stand+4, laneMax), "dirt")
		},
	},
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	console, err := rcon.Dial(address, password)
	if err != nil {
		return err
	}
	defer func() {
		_ = console.Close()
	}()

	b := &builder{console: console}

	b.run("gamerule doDaylightCycle false")
	b.run("gamerule doWeatherCycle false")
	b.run("gamerule randomTickSpeed 0")
	b.run("time set day")

	carve(b)
	for _, current := range stations {
		current.build(b)
	}
	reset(b)

	if b.err != nil {
		return b.err
	}

	fmt.Println("course rebuilt:")
	for _, current := range stations {
		fmt.Printf("  %-52s goal %v\n", current.name, current.goal)
	}
	fmt.Println("\ngoal flag:")
	fmt.Println(" ", goalFlag())

	return nil
}

// the corridor is sealed on all four sides: an open end lets the planner leave
// the arena and walk around the outside, which quietly skips every obstacle
func carve(b *builder) {
	b.fill(
		gocraft.At(courseStart, stand, -fenceZ),
		gocraft.At(courseEnd, ceiling, fenceZ),
		"air")
	b.fill(
		gocraft.At(courseStart, worldFloor+1, -fenceZ),
		gocraft.At(courseEnd, ground, fenceZ),
		"dirt")
	b.slab(courseStart, courseEnd, worldFloor, "bedrock")

	b.slab(courseStart, courseEnd, ground, "grass_block")
	b.barrier(courseStart, courseEnd, -fenceZ)
	b.barrier(courseStart, courseEnd, fenceZ)

	b.fill(
		gocraft.At(courseStart, stand, -fenceZ),
		gocraft.At(courseStart, ceiling, fenceZ),
		"bedrock")
	b.fill(
		gocraft.At(courseEnd, stand, -fenceZ),
		gocraft.At(courseEnd, ceiling, fenceZ),
		"bedrock")
}

func reset(b *builder) {
	entrance := gocraft.At(courseStart+3, stand, 0)

	b.run(fmt.Sprintf("setworldspawn %d %d %d", entrance.X, entrance.Y, entrance.Z))
	b.run(fmt.Sprintf("spawnpoint @a %d %d %d", entrance.X, entrance.Y, entrance.Z))
	b.run(fmt.Sprintf("tp @a %d %d %d", entrance.X, entrance.Y, entrance.Z))
}

func goalFlag() string {
	legs := make([]string, 0, len(stations))
	for _, current := range stations {
		legs = append(legs, fmt.Sprintf("%d,%d,%d", current.goal.X, current.goal.Y, current.goal.Z))
	}

	return strings.Join(legs, ";")
}
