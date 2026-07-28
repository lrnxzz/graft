package main

import "fmt"

// a protocol number pins exactly one Minecraft release, and the block model
// dataset published upstream trails it, so the two are recorded side by side
// instead of being spelled out again wherever a download needs them
type release struct {
	game   string
	models string
}

var releases = map[string]release{
	"765": {
		game:   "1.20.4",
		models: "1.20.2",
	},
}

func releaseOf(protocol string) (release, error) {
	known, listed := releases[protocol]
	if !listed {
		return release{}, fmt.Errorf("gen: protocol %s has no known Minecraft release", protocol)
	}

	return known, nil
}

func (r release) modelsURL() string {
	return fmt.Sprintf("https://raw.githubusercontent.com/PrismarineJS/minecraft-assets/master/data/%s/blocks_models.json", r.models)
}
