package plugin

// A Notice is something that already happened, and an Intent is something about
// to. Both carry their own payload as a struct rather than a bag of strings, so
// the host filling one and the declaration a plugin reads cannot drift apart.
type Notice interface {
	Event() string
}

type Intent interface {
	Intent() string
}

type Spawned struct {
	At Vec3 `json:"at"`
}

func (Spawned) Event() string {
	return "spawned"
}

type Arrived struct {
	At     Vec3   `json:"at"`
	Reason string `json:"reason"`
}

func (Arrived) Event() string {
	return "arrived"
}

type Said struct {
	Text string `json:"text"`
}

func (Said) Event() string {
	return "chat"
}

type BlockChanged struct {
	At    Vec3  `json:"at"`
	Block Block `json:"block"`
}

func (BlockChanged) Event() string {
	return "blockChanged"
}

type HealthChanged struct {
	Health float32 `json:"hp"`
	Food   float32 `json:"food"`
}

func (HealthChanged) Event() string {
	return "health"
}

type Disconnected struct {
	Reason string `json:"reason"`
}

func (Disconnected) Event() string {
	return "disconnected"
}

type Digging struct {
	Block Vec3 `json:"block"`
	Tool  Item `json:"tool"`
}

func (Digging) Intent() string {
	return "dig"
}

type Placing struct {
	Block Vec3 `json:"block"`
	Item  Item `json:"item"`
}

func (Placing) Intent() string {
	return "place"
}

type Chatting struct {
	Text string `json:"text"`
}

func (Chatting) Intent() string {
	return "chat"
}

// a navigation intent carries nothing: the goal it is about to chase knows only
// how to answer whether a position reaches it, so there is no destination to name
type Navigating struct{}

func (Navigating) Intent() string {
	return "move"
}

// Notices and Intents are what a plugin may subscribe to. They exist so the
// names live in one place: the runtime rejects anything else, and the test that
// guards graft.d.ts reads them rather than a second list.
func Notices() []Notice {
	return []Notice{
		Spawned{},
		Arrived{},
		Said{},
		BlockChanged{},
		HealthChanged{},
		Disconnected{},
	}
}

func Intents() []Intent {
	return []Intent{
		Digging{},
		Placing{},
		Chatting{},
		Navigating{},
	}
}
