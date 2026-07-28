package main

import (
	"context"
	"fmt"
	"time"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/agent"
	"github.com/lrnxzz/go-craft/host"
	"github.com/lrnxzz/go-craft/pathfinder"
)

const (
	address    = "localhost:25565"
	retryDelay = 3 * time.Second
)

// the stations built by probe/course, walked in order and then back to the
// entrance so the lap closes
var course = [...]gocraft.Position{
	gocraft.At(6, -60, 0),
	gocraft.At(16, -57, 0),
	gocraft.At(24, -60, 0),
	gocraft.At(34, -60, 0),
	gocraft.At(42, -60, 0),
	gocraft.At(50, -60, 0),
	gocraft.At(58, -60, 0),
	gocraft.At(70, -60, 0),
	gocraft.At(78, -60, 0),
	gocraft.At(-3, -60, 0),
}

func main() {
	fmt.Println("gocraft pathfinder demo — connecting to", address)

	for {
		if err := run(); err != nil {
			fmt.Println("session ended:", err)
		}

		fmt.Println("reconnecting in", retryDelay)
		time.Sleep(retryDelay)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	laps := func(ctx context.Context, bot *agent.Agent) error {
		for lap := 1; ; lap++ {
			fmt.Printf("=== lap %d ===\n", lap)

			for _, goal := range course {
				fmt.Printf("heading to %v from %v\n", goal, bot.Snapshot().Position)

				arrived, err := bot.Navigate(ctx, pathfinder.GoalAt(goal))
				if err != nil {
					fmt.Println("leg failed:", err)
				} else {
					fmt.Println("arrived at", arrived)
				}

				// a dropped session fails every leg from here on, so the lap
				// ends rather than spinning on a dead connection
				if ctx.Err() != nil {
					return ctx.Err()
				}

				time.Sleep(time.Second)
			}
		}
	}

	return host.Run(ctx, address, "gocraft_bot", laps)
}
