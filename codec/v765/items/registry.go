package items

import (
	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/codec/v765/assets"
	"github.com/lrnxzz/graft/lib"
)

//go:generate go run github.com/lrnxzz/graft/cli gen items 765

var registry = lib.LoadRegistry[graft.Item](765, assets.Items)

var Of = lib.Keyed(registry, func(i graft.Item) graft.ItemID {
	return i.ID
})

var Named = lib.Keyed(registry, func(i graft.Item) graft.Identifier {
	return i.Name
})

var Names = lib.Listed(registry, func(e graft.Item) graft.Identifier {
	return e.Name
})
