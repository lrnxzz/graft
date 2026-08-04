package plugin

type Marker struct {
	Kind  MarkerKind
	From  Vec3
	To    Vec3
	Color Color
}

type MarkerKind string

const (
	MarkHighlight MarkerKind = "highlight"
	MarkBox       MarkerKind = "box"
	MarkLine      MarkerKind = "line"
	MarkBeacon    MarkerKind = "beacon"
)

type MarkerSpec struct {
	Kind  MarkerKind
	Build func(args []reading) Marker
}

func Markers() []MarkerSpec {
	return []MarkerSpec{
		{Kind: MarkHighlight, Build: markSpot(MarkHighlight)},
		{Kind: MarkBeacon, Build: markSpot(MarkBeacon)},
		{Kind: MarkBox, Build: markSpan(MarkBox)},
		{Kind: MarkLine, Build: markSpan(MarkLine)},
	}
}

func markSpot(kind MarkerKind) func([]reading) Marker {
	return func(args []reading) Marker {
		return Marker{
			Kind:  kind,
			From:  vec3Of(argument(args, 0)),
			Color: argument(args, 1).color().or(white),
		}
	}
}

func markSpan(kind MarkerKind) func([]reading) Marker {
	return func(args []reading) Marker {
		return Marker{
			Kind:  kind,
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
