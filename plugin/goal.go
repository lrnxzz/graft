package plugin

type Goal struct {
	Kind   GoalKind
	At     Vec3
	Radius float64
	Ore    Block
	Item   Item
	Count  int
	Player Player
	Times  int
	Inner  []Goal
}

type GoalKind string

const (
	GoalAt       GoalKind = "at"
	GoalNear     GoalKind = "near"
	GoalMine     GoalKind = "mine"
	GoalCollect  GoalKind = "collect"
	GoalFollow   GoalKind = "follow"
	GoalFlee     GoalKind = "flee"
	GoalSequence GoalKind = "sequence"
	GoalRepeat   GoalKind = "repeat"
	GoalRace     GoalKind = "race"
)

type GoalSpec struct {
	Kind  GoalKind
	Build func(reading, reading) Goal
}

func Goals() []GoalSpec {
	return []GoalSpec{
		{Kind: GoalAt, Build: spotGoal(GoalAt)},
		{Kind: GoalNear, Build: spotGoal(GoalNear)},
		{Kind: GoalFlee, Build: spotGoal(GoalFlee)},
		{Kind: GoalMine, Build: minedGoal(GoalMine)},
		{Kind: GoalFollow, Build: followGoal(GoalFollow)},
		{Kind: GoalCollect, Build: countedGoal(GoalCollect)},
	}
}

func spotGoal(kind GoalKind) func(reading, reading) Goal {
	return func(first, second reading) Goal {
		return Goal{
			Kind:   kind,
			At:     vec3Of(first),
			Radius: second.decimal(),
		}
	}
}

func minedGoal(kind GoalKind) func(reading, reading) Goal {
	return func(first, second reading) Goal {
		return Goal{
			Kind:   kind,
			Ore:    Block(first.text()),
			Radius: second.field("radius").decimal(),
		}
	}
}

func followGoal(kind GoalKind) func(reading, reading) Goal {
	return func(first, second reading) Goal {
		return Goal{
			Kind:   kind,
			Player: Player(first.text()),
			Radius: second.field("distance").decimal(),
		}
	}
}

func countedGoal(kind GoalKind) func(reading, reading) Goal {
	return func(first, second reading) Goal {
		return Goal{
			Kind:  kind,
			Item:  Item(first.text()),
			Count: second.count(),
		}
	}
}

func asGoal(given reading) Goal {
	goal, _ := exported[Goal](given)

	return goal
}
