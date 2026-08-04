package plugin

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

type MarkerSpec struct {
	Type  MarkerType
	Build func(args []reading) Marker
}

func Markers() []MarkerSpec {
	return []MarkerSpec{
		{Type: MarkHighlight, Build: markSpot(MarkHighlight)},
		{Type: MarkBeacon, Build: markSpot(MarkBeacon)},
		{Type: MarkBox, Build: markSpan(MarkBox)},
		{Type: MarkLine, Build: markSpan(MarkLine)},
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
