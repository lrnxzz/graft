package plugin

type box struct {
	x      float32
	y      float32
	width  float32
	height float32
}

func (b box) contains(x, y float32) bool {
	return x >= b.x && x < b.x+b.width && y >= b.y && y < b.y+b.height
}

type Pick struct {
	at box
	on func()
}

func (p Pick) Bounds() (x, y, width, height float32) {
	return p.at.x, p.at.y, p.at.width, p.at.height
}

type placement struct {
	surface Surface
	scroll  *Scroll
	picks   []Pick
}

type Scroll struct {
	Offset float32
	extent float32
}

func (s *Scroll) By(delta float32) {
	if s == nil {
		return
	}

	s.Offset = min(max(s.Offset-delta, 0), s.extent)
}

func (s *Scroll) at() float32 {
	if s == nil {
		return 0
	}

	return s.Offset
}

func (s *Scroll) fits(content, height float32) {
	if s == nil {
		return
	}

	s.extent = max(content-height, 0)
	s.Offset = min(s.Offset, s.extent)
}

func Clicked(picks []Pick, x, y float32) func() {
	var hit func()
	for _, candidate := range picks {
		if candidate.at.contains(x, y) {
			hit = candidate.on
		}
	}

	return hit
}

func anchorAt(anchor Anchor, screenWidth, screenHeight, width, height float32) (float32, float32) {
	var x, y float32

	switch anchor {
	case TopLeft, Left, BottomLeft, "":
		x = 0
	case Top, Center, Bottom:
		x = (screenWidth - width) / 2
	default:
		x = screenWidth - width
	}

	switch anchor {
	case TopLeft, Top, TopRight:
		y = 0
	case Left, Center, Right:
		y = (screenHeight - height) / 2
	default:
		y = screenHeight - height
	}

	return x, y
}

func Paint(root Node, surface Surface) []Pick {
	return Scrolled(root, surface, nil)
}

func Scrolled(root Node, surface Surface, scroll *Scroll) []Pick {
	if root == nil {
		return nil
	}

	spot := &placement{
		surface: surface,
		scroll:  scroll,
	}
	width, height := root.Measure(spot)

	screenWidth, screenHeight := surface.Size()
	x, y := anchorAt(anchorOf(root), screenWidth, screenHeight, width, height)

	root.Paint(spot, box{
		x:      x,
		y:      y,
		width:  width,
		height: height,
	})

	return spot.picks
}

func anchorOf(node Node) Anchor {
	anchored, placed := node.(Anchored)
	if !placed {
		return TopLeft
	}

	return anchored.Anchor()
}
