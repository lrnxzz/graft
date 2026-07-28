package viewer

import "github.com/lrnxzz/go-craft/viewer/gpu"

// Layer draws over the world, in screen space. Everything the viewer itself puts
// on screen — the hud, the chat, the crosshair — is a Layer, so a plugin's is not
// a second-class citizen.
type Layer interface {
	Draw(*Canvas)
}

// WorldLayer draws inside the world, before the overlay and with depth still on.
// A type may be both; the viewer walks the two phases separately.
type WorldLayer interface {
	DrawWorld(*Painter)
}

// Screen is a menu. While one is on the stack the bot stops taking movement
// input, the cursor is released, and clicks and keys reach the topmost screen
// instead of the game.
type Screen interface {
	Draw(*Canvas)
	Click(cursor gpu.Point, screen gpu.Rect)
	Key(key gpu.Key)
}

// Closer lets a Layer or Screen release GPU resources when the viewer shuts down
// or the screen is dismissed. Implementing it is optional.
type Closer interface {
	Close()
}

func closeIfPossible(value any) {
	closer, ok := value.(Closer)
	if ok {
		closer.Close()
	}
}
