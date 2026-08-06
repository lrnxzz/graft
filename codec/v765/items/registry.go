package items

import (
	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/codec/v765/assets"
	"github.com/lrnxzz/go-craft/lib"
)

//go:generate go run github.com/lrnxzz/go-craft/cli gen items 765

var registry = lib.LoadRegistry[gocraft.Item](765, assets.Items)

var Of = lib.Keyed(registry, func(i gocraft.Item) gocraft.ItemID {
	return i.ID
})

var Named = lib.Keyed(registry, func(i gocraft.Item) gocraft.Identifier {
	return i.Name
})

var Names = lib.Listed(registry, func(e gocraft.Item) gocraft.Identifier {
	return e.Name
})
