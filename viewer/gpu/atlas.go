package gpu

type UV struct {
	U0, V0, U1, V1 float32
}

type Atlas struct {
	texture *Texture
	columns int
	rows    int
}

func NewAtlas(texture *Texture, columns, rows int) *Atlas {
	return &Atlas{texture: texture, columns: columns, rows: rows}
}

func (a *Atlas) Bind(unit uint32) {
	a.texture.Bind(unit)
}

func (a *Atlas) Texture() *Texture {
	return a.texture
}

func (a *Atlas) Delete() {
	a.texture.Delete()
}

func (a *Atlas) Tile(index int) UV {
	column := float32(index % a.columns)
	row := float32(index / a.columns)

	return UV{
		U0: column / float32(a.columns),
		V0: row / float32(a.rows),
		U1: (column + 1) / float32(a.columns),
		V1: (row + 1) / float32(a.rows),
	}
}
