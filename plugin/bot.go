package plugin

type Bot interface {
	Name() string
	Position() Vec3
}

type Vitals interface {
	Health() float32
	Food() float32
	OnGround() bool
}

type Sighted interface {
	Looking() (Target, bool)
	BlockAt(at Vec3) Block
}

type Holder interface {
	Held() Item
	Inventory() []Stack
	Count(item Item) int
	Hold(item Item) error
}

type Walker interface {
	Goto(at Vec3) (Vec3, error)
	Look(at Vec3)
}

type Pursuer interface {
	Pursue(goal Goal) error
	Abandon()
}

type Digger interface {
	Dig(at Vec3) error
}

type Builder interface {
	Place(at Vec3) error
}

type Speaker interface {
	Say(text string) error
}

type Watcher interface {
	Watch(event string, handle func(Notice)) bool
}

type Guard interface {
	Guard(intent string, handle func(Intent) string) bool
}

type Target struct {
	At    Vec3  `json:"at"`
	Face  Face  `json:"face"`
	Block Block `json:"block"`
}

type Stack struct {
	Item  Item `json:"item"`
	Count int  `json:"count"`
	Slot  int  `json:"slot"`
}
