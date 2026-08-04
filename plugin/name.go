package plugin

type (
	Block  string
	Item   string
	Player string
)

type Face string

const (
	Down  Face = "down"
	Up    Face = "up"
	North Face = "north"
	South Face = "south"
	West  Face = "west"
	East  Face = "east"
)

type Sprite string

func (b Block) Sprite() Sprite {
	return Sprite(b)
}

func (i Item) Sprite() Sprite {
	return Sprite(i)
}
