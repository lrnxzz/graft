package main

import (
	"context"
	"fmt"
	"log"
	"time"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/agent"
	"github.com/lrnxzz/go-craft/host"
	"github.com/lrnxzz/go-craft/pathfinder"
)

var legs = [...]gocraft.Position{
	gocraft.At(92, -55, 133),
	gocraft.At(71, -57, 119),
}

const address = "localhost:25565"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	username := fmt.Sprintf("scan_%d", time.Now().Unix()%100000)

	scan := func(ctx context.Context, bot *agent.Agent) error {
		for _, leg := range legs {
			fmt.Printf("navigating to %v from %v\n", leg, bot.Snapshot().Position)

			arrived, err := bot.Navigate(ctx, pathfinder.GoalAt(leg))
			fmt.Printf("leg result: arrived=%v err=%v at=%v\n", arrived, err, bot.Snapshot().Position)
			if err != nil {
				return err
			}
		}

		return nil
	}

	return host.Run(ctx, address, username, scan)
}
