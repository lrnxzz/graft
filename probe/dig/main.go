package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/lrnxzz/go-craft/agent"
	"github.com/lrnxzz/go-craft/host"
)

const (
	address = "localhost:25565"
	reach   = 4.5
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	excavate := func(ctx context.Context, bot *agent.Agent) error {
		bot.Look(0, 75)
		time.Sleep(300 * time.Millisecond)

		hit, sighted := bot.Target(reach)
		if !sighted {
			return errors.New("probe: no block in the crosshair")
		}
		before, _ := bot.World().BlockAt(hit.Block)
		fmt.Printf("aiming at %v (state %d)\n", hit.Block, before)

		dug, err := bot.Dig(ctx, reach)
		if err != nil {
			return fmt.Errorf("probe: dig: %w", err)
		}

		time.Sleep(time.Second)
		after, _ := bot.World().BlockAt(dug.Block)
		fmt.Printf("broke %v: state before=%d after=%d\n", dug.Block, before, after)

		if after == 0 {
			fmt.Println("DIG OK: the block turned into air in the world")

			return nil
		}

		fmt.Println("DIG FAILED: the block is still there")

		return nil
	}

	return host.Run(ctx, address, "gocraft_dig", excavate)
}
