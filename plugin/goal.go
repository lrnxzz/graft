package plugin

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

type GoalSpec struct {
	Type  GoalType
	Build func(reading, reading) Goal
}

func Goals() []GoalSpec {
	return []GoalSpec{
		{
			Type:  GoalAt,
			Build: spotGoal(GoalAt),
		},
		{
			Type:  GoalNear,
			Build: spotGoal(GoalNear),
		},
		{
			Type:  GoalFlee,
			Build: spotGoal(GoalFlee),
		},
		{
			Type:  GoalMine,
			Build: minedGoal(GoalMine),
		},
		{
			Type:  GoalFollow,
			Build: followGoal(GoalFollow),
		},
		{
			Type:  GoalCollect,
			Build: countedGoal(GoalCollect),
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
