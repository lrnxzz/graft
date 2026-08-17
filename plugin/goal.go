package plugin

import "github.com/dop251/goja"

type Goal struct {
	Type   GoalType
	At     Vec3
	Radius float64
	Ore    Block
	Item   Item
	Count  int
	Player Player
	Times  int
	Inner  []Goal
}

type GoalType string

const (
	GoalAt       GoalType = "at"
	GoalNear     GoalType = "near"
	GoalMine     GoalType = "mine"
	GoalCollect  GoalType = "collect"
	GoalFollow   GoalType = "follow"
	GoalFlee     GoalType = "flee"
	GoalSequence GoalType = "sequence"
	GoalRepeat   GoalType = "repeat"
	GoalRace     GoalType = "race"
)

func goalWords() []Word {
	return []Word{
		goalWord(GoalAt, spotGoal(GoalAt)),
		goalWord(GoalNear, spotGoal(GoalNear)),
		goalWord(GoalFlee, spotGoal(GoalFlee)),
		goalWord(GoalMine, minedGoal(GoalMine)),
		goalWord(GoalFollow, followGoal(GoalFollow)),
		goalWord(GoalCollect, countedGoal(GoalCollect)),
		nesting(GoalSequence),
		nesting(GoalRace),
		repeating(),
		area(),
	}
}

// goalWord reads the two arguments every simple goal takes
func goalWord(kind GoalType, build func(reading, reading) Goal) Word {
	return Word{
		Name: string(kind),
		Build: func(r *Runtime, call goja.FunctionCall) goja.Value {
			return r.vm.ToValue(build(r.reading(call.Argument(0)), r.reading(call.Argument(1))))
		},
	}
}

// nesting is a combinator holding any number of inner goals
func nesting(kind GoalType) Word {
	return Word{
		Name: string(kind),
		Build: func(r *Runtime, call goja.FunctionCall) goja.Value {
			goal := Goal{Type: kind}
			for _, argument := range call.Arguments {
				goal.Inner = append(goal.Inner, asGoal(r.reading(argument)))
			}

			return r.vm.ToValue(goal)
		},
	}
}

func repeating() Word {
	return Word{
		Name: string(GoalRepeat),
		Build: func(r *Runtime, call goja.FunctionCall) goja.Value {
			return r.vm.ToValue(Goal{
				Type:  GoalRepeat,
				Times: r.reading(call.Argument(1)).count(),
				Inner: []Goal{asGoal(r.reading(call.Argument(0)))},
			})
		},
	}
}

// area is what mine takes as its reach: within(radius)
func area() Word {
	return Word{
		Name: "within",
		Build: func(r *Runtime, call goja.FunctionCall) goja.Value {
			area := r.vm.NewObject()
			_ = area.Set("radius", call.Argument(0))

			return area
		},
	}
}

func spotGoal(wanted GoalType) func(reading, reading) Goal {
	return func(first, second reading) Goal {
		return Goal{
			Type:   wanted,
			At:     vec3Of(first),
			Radius: second.decimal(),
		}
	}
}

func minedGoal(wanted GoalType) func(reading, reading) Goal {
	return func(first, second reading) Goal {
		return Goal{
			Type:   wanted,
			Ore:    Block(first.text()),
			Radius: second.field("radius").decimal(),
		}
	}
}

func followGoal(wanted GoalType) func(reading, reading) Goal {
	return func(first, second reading) Goal {
		return Goal{
			Type:   wanted,
			Player: Player(first.text()),
			Radius: second.field("distance").decimal(),
		}
	}
}

func countedGoal(wanted GoalType) func(reading, reading) Goal {
	return func(first, second reading) Goal {
		return Goal{
			Type:  wanted,
			Item:  Item(first.text()),
			Count: second.count(),
		}
	}
}

func asGoal(given reading) Goal {
	goal, _ := exported[Goal](given)

	return goal
}
