package pathfinder

import "math"

// every cost is denominated in blocks walked, so a tick spent mining is worth
// the ground the bot would have covered walking during that same tick — that is
// what lets the search weigh a tunnel against the detour around it
const (
	costWalk     = 1.0
	costDiagonal = math.Sqrt2
	costClimb    = 2.0
	costFall     = 0.5
	costParkour  = 3.0
	costPlace    = 2.0
	costTick     = 0.21
)
