package command

import (
	"math"
	"strconv"
	"strings"

	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/codec/v765/blocks"
	"github.com/lrnxzz/graft/codec/v765/items"
)

const namespace = "minecraft:"

// Spot is three coordinates. Each may be absolute, relative to where the bot
// stands with ~, or relative to where it faces with ^ — the three forms players
// already know, so nobody has to learn ours.
func Spot(label string) Param[graft.Position] {
	read := func(stream *Stream, call Call) (graft.Position, error) {
		axes, local, err := coordinates(label, stream)
		if err != nil {
			return graft.Position{}, err
		}

		if local {
			return ahead(call, axes), nil
		}

		return graft.Position{
			X: int(axes[0]),
			Y: int(axes[1]),
			Z: int(axes[2]),
		}, nil
	}

	// resting the aim on a block is the fastest way to name one, so that is what
	// an empty coordinate offers
	offer := func(call Call, typed string) []string {
		if typed != "" || call.bot == nil {
			return nil
		}

		hit, sighted := call.bot.Target(graft.BlockReach)
		if !sighted {
			return nil
		}

		return []string{strconv.Itoa(hit.Block.X) + " " + strconv.Itoa(hit.Block.Y) + " " + strconv.Itoa(hit.Block.Z)}
	}

	return Param[graft.Position]{
		label: label,
		shape: "x y z",
		read:  read,
		offer: offer,
	}
}

// coordinates reads the three axes, allowing commas between them. Local offsets
// are all or nothing: ^ mixed with a plain number has no meaning, because the
// three local axes are read off one basis.
func coordinates(label string, stream *Stream) ([3]float64, bool, error) {
	var (
		axes  [3]float64
		local bool
		here  bool
	)

	for axis := range 3 {
		if axis > 0 {
			stream.Skip(Comma)
		}

		token, held := stream.Take()
		if !held {
			return axes, false, missing(label)
		}

		switch token.Shape {
		case Numeral:
			value, err := strconv.ParseFloat(token.Text, 64)
			if err != nil {
				return axes, false, refuse(label, token, "a coordinate")
			}

			axes[axis] = value
		case Relative, Local:
			here = true
			local = local || token.Shape == Local

			offset, carried := token.Offset()
			if !carried {
				continue
			}

			value, err := strconv.ParseFloat(offset, 64)
			if err != nil {
				return axes, false, refuse(label, token, "an offset such as ~5 or ^-2")
			}

			axes[axis] = value
		default:
			return axes, false, refuse(label, token, "a coordinate — a number, ~ from here or ^ from where the bot faces")
		}

		if token.Shape == Local && !local {
			local = true
		}
	}

	if local && !here {
		return axes, false, nil
	}

	return axes, local, nil
}

// ahead turns local offsets into a block: left, up and forward from where the
// bot is looking, which is what ^ means to a player
func ahead(call Call, axes [3]float64) graft.Position {
	standing := call.Standing()
	if call.bot == nil {
		return standing
	}

	snapshot := call.bot.Snapshot()

	yaw := float64(snapshot.Yaw) * math.Pi / 180
	pitch := float64(snapshot.Pitch) * math.Pi / 180

	forward := graft.Vec3(-math.Sin(yaw)*math.Cos(pitch), -math.Sin(pitch), math.Cos(yaw)*math.Cos(pitch))
	left := graft.Vec3(math.Cos(yaw), 0, math.Sin(yaw))
	up := graft.Vec3(0, 1, 0)

	at := standing.Center().
		Add(left.Scale(axes[0])).
		Add(up.Scale(axes[1])).
		Add(forward.Scale(axes[2]))

	return at.Floor()
}

// Block is a block name checked against the registry, so a command refuses a
// name the world never had instead of walking off after it
func Block(label string) Param[graft.Block] {
	read := func(stream *Stream, _ Call) (graft.Block, error) {
		token, held := stream.Take()
		if !held {
			return graft.Block{}, missing(label)
		}

		found, known := lookup(token.Text, blocks.Named)
		if known {
			return found, nil
		}

		return graft.Block{}, unknown(label, token, "block", blocks.Names)
	}

	offer := func(_ Call, typed string) []string {
		return suggest(blocks.Names, typed)
	}

	return Param[graft.Block]{
		label: label,
		shape: "block",
		read:  read,
		offer: offer,
	}
}

func Item(label string) Param[graft.Item] {
	read := func(stream *Stream, _ Call) (graft.Item, error) {
		token, held := stream.Take()
		if !held {
			return graft.Item{}, missing(label)
		}

		found, known := lookup(token.Text, items.Named)
		if known {
			return found, nil
		}

		return graft.Item{}, unknown(label, token, "item", items.Names)
	}

	offer := func(_ Call, typed string) []string {
		return suggest(items.Names, typed)
	}

	return Param[graft.Item]{
		label: label,
		shape: "item",
		read:  read,
		offer: offer,
	}
}

// a player may write either form, and the registry is keyed on whichever the
// data happened to carry, so both are tried rather than guessed at
func lookup[T any](name string, find func(graft.Identifier) (T, bool)) (T, bool) {
	written := graft.Identifier(name)

	found, known := find(written)
	if known {
		return found, true
	}

	return find(graft.Identifier(written.Path()))
}

// unknown reports a name the registry never had, with the nearest one it does,
// because a typo in a thousand names is otherwise a guessing game
func unknown(label string, token Token, sort string, names []graft.Identifier) error {
	refused := &Refusal{
		Label:  label,
		Got:    token.Text,
		Wanted: "a " + sort + " this world knows",
		From:   token.From,
		To:     token.To,
	}

	near := suggest(names, token.Text)
	if len(near) > 0 {
		refused.Near = near[0]
	}

	return refused
}

const offered = 12

// suggest ranks by prefix first and then by containment, so typing the middle of
// a name still finds it while the obvious match stays on top
func suggest(names []graft.Identifier, typed string) []string {
	wanted := strings.ToLower(strings.TrimPrefix(typed, namespace))

	var (
		starting []string
		holding  []string
	)

	for _, name := range names {
		path := name.Path()

		switch {
		case wanted == "" || strings.HasPrefix(path, wanted):
			starting = append(starting, path)
		case strings.Contains(path, wanted):
			holding = append(holding, path)
		}

		if len(starting) >= offered {
			break
		}
	}

	found := append(starting, holding...)
	if len(found) > offered {
		found = found[:offered]
	}

	return found
}
