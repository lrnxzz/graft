package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/agent"
	"github.com/lrnxzz/go-craft/host"
	"github.com/lrnxzz/go-craft/pathfinder"
)

const address = "localhost:25565"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	legs := make([]gocraft.Position, 0, len(os.Args)-1)
	for _, arg := range os.Args[1:] {
		leg, err := gocraft.ParsePosition(arg)
		if err != nil {
			return fmt.Errorf("probe: bad leg %q: %w", arg, err)
		}

		legs = append(legs, leg)
	}

	course := func(ctx context.Context, bot *agent.Agent) error {
		for _, leg := range legs {
			fmt.Printf("navigating to %v from %v\n", leg, bot.Snapshot().Position)

			if err := march(ctx, bot, leg); err != nil {
				return err
			}
		}
		fmt.Println("all legs complete")

		return nil
	}

	return host.Run(ctx, address, "gocraft_live", course)
}

func march(ctx context.Context, bot *agent.Agent, leg gocraft.Position) error {
	walking, stop := context.WithCancel(ctx)
	defer stop()

	go trace(walking, bot)

	arrived, err := bot.Navigate(ctx, pathfinder.GoalAt(leg))
	fmt.Printf("leg result: arrived=%v err=%v at=%v\n", arrived, err, bot.Snapshot().Position)

	return err
}

// trace reports where the bot is once a second until its leg ends
func trace(ctx context.Context, bot *agent.Agent) {
	progress := time.NewTicker(time.Second)
	defer progress.Stop()

	for {
		select {
		case <-progress.C:
			snapshot := bot.Snapshot()
			fmt.Printf("  walking at=%v ground=%v\n", snapshot.Position, snapshot.OnGround)
		case <-ctx.Done():
			return
		}
	}
}
