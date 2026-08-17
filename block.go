package graft

import (
	"encoding/json"
	"strconv"

	"github.com/lrnxzz/graft/codec"
)

type BlockState int32

var blockPalette = codec.PaletteType{
	Entries:     4096,
	IndirectMax: 8,
	DirectBits:  15,
}

func BlockStates() codec.PalettedContainer[BlockState] {
	return codec.PalettedContainer[BlockState]{
		PaletteType: blockPalette,
	}
}

type Property struct {
	Name   string
	Values []string
}

func (p *Property) UnmarshalJSON(data []byte) error {
	var raw propertyData
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	p.Name = raw.Name
	switch raw.Type {
	case "bool":
		p.Values = []string{
			"true",
			"false",
		}
	case "enum":
		p.Values = raw.Values
	default:
		p.Values = make([]string, raw.NumValues)
		for i := range p.Values {
			p.Values[i] = strconv.Itoa(i)
		}
	}

	return nil
}

type propertyData struct {
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	NumValues int      `json:"num_values"`
	Values    []string `json:"values"`
}

type BoundingBox string

const (
	BoundingBoxEmpty BoundingBox = "empty"
	BoundingBoxBlock BoundingBox = "block"
)

type Block struct {
	Name         Identifier      `json:"name"`
	MinState     BlockState      `json:"minStateId"`
	MaxState     BlockState      `json:"maxStateId"`
	DefaultState BlockState      `json:"defaultState"`
	BoundingBox  BoundingBox     `json:"boundingBox"`
	Hardness     float64         `json:"hardness"`
	Diggable     bool            `json:"diggable"`
	Material     string          `json:"material"`
	HarvestTools map[ItemID]bool `json:"harvestTools"`
	Properties   []Property      `json:"states"`
}

func (b Block) Solid() bool {
	return b.BoundingBox == BoundingBoxBlock
}

func (b Block) At(state BlockState) map[string]string {
	values := make(map[string]string, len(b.Properties))

	offset := int(state - b.MinState)
	for i := len(b.Properties) - 1; i >= 0; i-- {
		property := b.Properties[i]
		values[property.Name] = property.Values[offset%len(property.Values)]
		offset /= len(property.Values)
	}

	return values
}
