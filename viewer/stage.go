package viewer

import "github.com/lrnxzz/go-craft/viewer/gpu"

// stage is everything the viewer hands out: drawing time to layers, the frame to
// menus, keys to whoever claimed them. The viewer's own hud and chat go through
// it too, so a plugin's are never a second kind of thing.
type stage struct {
	world   []WorldLayer
	layers  []Layer
	screens []Screen
	binds   map[gpu.Key]func()
	spoken  []func(string) bool
}

func newStage() stage {
	return stage{
		binds: map[gpu.Key]func(){},
	}
}

func (s *stage) addWorldLayer(layer WorldLayer) {
	s.world = append(s.world, layer)
}

func (s *stage) addLayer(layer Layer) {
	s.layers = append(s.layers, layer)
}

func (s *stage) bind(key gpu.Key, action func()) {
	s.binds[key] = action
}

// intercept offers a submitted line to each handler in turn. The first to answer
// true has dealt with it, and the server never sees it.
func (s *stage) intercept(handle func(string) bool) {
	s.spoken = append(s.spoken, handle)
}

func (s *stage) claimed(line string) bool {
	for _, handle := range s.spoken {
		if handle(line) {
			return true
		}
	}

	return false
}

func (s *stage) open(screen Screen) {
	s.screens = append(s.screens, screen)
}

// dismiss reports whether the last menu just left, which is when the game takes
// the cursor back
func (s *stage) dismiss() bool {
	if len(s.screens) == 0 {
		return false
	}

	closeIfPossible(s.screens[len(s.screens)-1])
	s.screens = s.screens[:len(s.screens)-1]

	return len(s.screens) == 0
}

func (s *stage) showing() bool {
	return len(s.screens) > 0
}

func (s *stage) top() Screen {
	if len(s.screens) == 0 {
		return nil
	}

	return s.screens[len(s.screens)-1]
}

func (s *stage) drawWorld(painter *Painter) {
	for _, layer := range s.world {
		layer.DrawWorld(painter)
	}
}

func (s *stage) draw(canvas *Canvas) {
	for _, layer := range s.layers {
		layer.Draw(canvas)
	}
	for _, screen := range s.screens {
		screen.Draw(canvas)
	}
}

func (s *stage) fire(edges *edges) bool {
	struck := false
	for key, action := range s.binds {
		if edges.key(key).started {
			action()
			struck = true
		}
	}

	return struck
}

func (s *stage) close() {
	for _, layer := range s.world {
		closeIfPossible(layer)
	}
	for _, layer := range s.layers {
		closeIfPossible(layer)
	}
	for range s.screens {
		s.dismiss()
	}
}
