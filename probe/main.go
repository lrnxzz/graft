package main

import (
	"context"
	"fmt"
	"log"
	"time"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/agent"
	"github.com/lrnxzz/go-craft/pathfinder"
)

var legs = [...]gocraft.Position{
	gocraft.At(92, -55, 133),
	gocraft.At(71, -57, 119),
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	username := fmt.Sprintf("scan_%d", time.Now().Unix()%100000)

	bot, err := agent.Join(ctx, "localhost:25565", username)
	if err != nil {
		return err
	}

	go func() {
		if err := bot.Run(ctx); err != nil {
			log.Println("run:", err)
		}
	}()

	if err := bot.Ready(ctx); err != nil {
		return err
	}
	time.Sleep(time.Second)

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
