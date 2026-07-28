package atlas

import gocraft "github.com/lrnxzz/go-craft"

// Sheet is how a generated atlas image is cut into tiles. It ships as JSON beside
// the PNG and lives here, in the one module both sides can import, so the
// generator that writes the file and the viewer that reads it cannot drift.
type Sheet struct {
	Tile    int `json:"tile"`
	Columns int `json:"columns"`
	Rows    int `json:"rows"`
}

// Faces are the three tiles a block's cube is drawn from
type Faces struct {
	Up   int `json:"up"`
	Down int `json:"down"`
	Side int `json:"side"`
}

// BlockSheet is keyed by the bare block name, the same form the block registry
// carries, so a lookup needs no conversion on either side
type BlockSheet struct {
	Sheet
	Blocks map[gocraft.Identifier]Faces `json:"blocks"`
}

// ItemSheet is keyed by item path, which is what the sprite files are named after
type ItemSheet struct {
	Sheet
	Items map[string]int `json:"items"`
}
