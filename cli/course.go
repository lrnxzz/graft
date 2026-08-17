package main

import (
	"fmt"
	"strings"

	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/rcon"
	"github.com/spf13/cobra"
)

const (
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

func courseCommand() *cobra.Command {
	var password string

	command := &cobra.Command{
		Use:   "course <host[:port]>",
		Short: "Carve the pathfinder obstacle course on a server over rcon",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return carveCourse(args[0], password)
		},
	}

	command.Flags().StringVar(&password, "password", "graft", "rcon password")

	return command
}

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

func (b *builder) fill(from, to graft.Position, block string) {
	b.run(fmt.Sprintf("fill %d %d %d %d %d %d %s",
		from.X, from.Y, from.Z, to.X, to.Y, to.Z, block))
}

func (b *builder) slab(x1, x2, y int, block string) {
	b.fill(graft.At(x1, y, laneMin), graft.At(x2, y, laneMax), block)
}

func (b *builder) barrier(x1, x2, z int) {
	b.fill(graft.At(x1, stand, z), graft.At(x2, ceiling, z), "bedrock")
}

type station struct {
	name  string
	goal  graft.Position
	build func(*builder)
}

// every station is a walled corridor: without the fences the planner simply
// walks around the obstacle and never has to mine or bridge anything
var stations = [...]station{
	{
		name: "flat ground",
		goal: graft.At(6, stand, 0),
		build: func(*builder) {
		},
	},
	{
		name: "stairs up three",
		goal: graft.At(16, stand+3, 0),
		build: func(b *builder) {
			for step := range 3 {
				b.slab(11+step*2, 12+step*2, ground+1+step, "stone_bricks")
			}
			b.slab(17, 20, ground+3, "stone_bricks")
		},
	},
	{
		name: "three block drop",
		goal: graft.At(24, stand, 0),
		build: func(b *builder) {
			b.slab(21, 21, ground+3, "stone_bricks")
		},
	},
	{
		name: "parkour gap",
		goal: graft.At(34, stand, 0),
		build: func(b *builder) {
			b.fill(graft.At(29, pitFloor, laneMin), graft.At(31, ground, laneMax), "air")
		},
	},
	{
		name: "short dirt wall: climbs over, breaking the cap",
		goal: graft.At(42, stand, 0),
		build: func(b *builder) {
			b.fill(graft.At(38, stand, laneMin), graft.At(38, stand+1, laneMax), "dirt")
		},
	},
	{
		name: "tall dirt wall: tunnels through at foot level",
		goal: graft.At(50, stand, 0),
		build: func(b *builder) {
			b.fill(graft.At(46, stand, laneMin), graft.At(46, stand+5, laneMax), "dirt")
		},
	},
	{
		name: "stone wall with a gap: walking around beats mining",
		goal: graft.At(58, stand, 0),
		build: func(b *builder) {
			b.fill(graft.At(54, stand, laneMin), graft.At(54, stand+3, 3), "stone")
		},
	},
	{
		name: "four block chasm: leaps at the limit of the jump",
		goal: graft.At(70, stand, 0),
		build: func(b *builder) {
			b.fill(graft.At(63, pitFloor, laneMin), graft.At(66, ground, laneMax), "air")
		},
	},
	{
		name: "buried chamber: digs a shaft down",
		goal: graft.At(78, stand, 0),
		build: func(b *builder) {
			b.fill(graft.At(74, stand, laneMin), graft.At(80, stand+3, laneMax), "dirt")
			b.fill(graft.At(77, stand, -1), graft.At(79, stand+1, 1), "air")
			b.fill(graft.At(74, stand+4, laneMin), graft.At(80, stand+4, laneMax), "dirt")
		},
	},
}

// entrance is where the course starts and where a lap closes
func entrance() graft.Position {
	return graft.At(courseStart+3, stand, 0)
}

func carveCourse(address, password string) error {
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
	legs := make([]string, 0, len(stations)+1)
	for _, at := range lap() {
		legs = append(legs, written(at))
	}

	fmt.Println("\nwalk it with:")
	fmt.Println("  graft goto <host[:port]>", strings.Join(legs, " "))

	return nil
}

// the corridor is sealed on all four sides: an open end lets the planner leave
// the arena and walk around the outside, which quietly skips every obstacle
func carve(b *builder) {
	b.fill(
		graft.At(courseStart, stand, -fenceZ),
		graft.At(courseEnd, ceiling, fenceZ),
		"air")
	b.fill(
		graft.At(courseStart, worldFloor+1, -fenceZ),
		graft.At(courseEnd, ground, fenceZ),
		"dirt")
	b.slab(courseStart, courseEnd, worldFloor, "bedrock")

	b.slab(courseStart, courseEnd, ground, "grass_block")
	b.barrier(courseStart, courseEnd, -fenceZ)
	b.barrier(courseStart, courseEnd, fenceZ)

	b.fill(
		graft.At(courseStart, stand, -fenceZ),
		graft.At(courseStart, ceiling, fenceZ),
		"bedrock")
	b.fill(
		graft.At(courseEnd, stand, -fenceZ),
		graft.At(courseEnd, ceiling, fenceZ),
		"bedrock")
}

func reset(b *builder) {
	at := entrance()

	b.run(fmt.Sprintf("setworldspawn %d %d %d", at.X, at.Y, at.Z))
	b.run(fmt.Sprintf("spawnpoint @a %d %d %d", at.X, at.Y, at.Z))
	b.run(fmt.Sprintf("tp @a %d %d %d", at.X, at.Y, at.Z))
}

// lap is every station in order and then the entrance, so the walk closes where
// it started. Both the demo and the goto line printed above read it, which is
// what keeps either from drifting from the course actually carved.
func lap() []graft.Position {
	legs := make([]graft.Position, 0, len(stations)+1)
	for _, current := range stations {
		legs = append(legs, current.goal)
	}

	return append(legs, entrance())
}

func written(at graft.Position) string {
	return fmt.Sprintf("%d,%d,%d", at.X, at.Y, at.Z)
}
