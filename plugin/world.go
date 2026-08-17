package plugin

import "github.com/dop251/goja"

type Marker struct {
	Type  MarkerType
	From  Vec3
	To    Vec3
	Color Color
}

type MarkerType string

const (
	MarkHighlight MarkerType = "highlight"
	MarkBox       MarkerType = "box"
	MarkLine      MarkerType = "line"
	MarkBeacon    MarkerType = "beacon"
)

func markerWords() []Word {
	return []Word{
		marker(MarkHighlight, markSpot(MarkHighlight)),
		marker(MarkBeacon, markSpot(MarkBeacon)),
		marker(MarkBox, markSpan(MarkBox)),
		marker(MarkLine, markSpan(MarkLine)),
	}
}

// marker gathers however many arguments the mark's shape wants
func marker(kind MarkerType, build func([]reading) Marker) Word {
	return Word{
		Name: string(kind),
		Build: func(r *Runtime, call goja.FunctionCall) goja.Value {
			args := make([]reading, 0, len(call.Arguments))
			for _, argument := range call.Arguments {
				args = append(args, r.reading(argument))
			}

			return r.vm.ToValue(build(args))
		},
	}
}

func markSpot(drawn MarkerType) func([]reading) Marker {
	return func(args []reading) Marker {
		return Marker{
			Type:  drawn,
			From:  vec3Of(argument(args, 0)),
			Color: argument(args, 1).color().or(white),
		}
	}
}

func markSpan(drawn MarkerType) func([]reading) Marker {
	return func(args []reading) Marker {
		return Marker{
			Type:  drawn,
			From:  vec3Of(argument(args, 0)),
			To:    vec3Of(argument(args, 1)),
			Color: argument(args, 2).color().or(white),
		}
	}
}

func argument(args []reading, index int) reading {
	if index >= len(args) {
		return reading{}
	}

	return args[index]
}

func vec3Of(given reading) Vec3 {
	at, is := exported[Vec3](given)
	if is {
		return at
	}

	return Vec3{
		X: given.field("x").decimal(),
		Y: given.field("y").decimal(),
		Z: given.field("z").decimal(),
	}
}
