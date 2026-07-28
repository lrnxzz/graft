package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/lrnxzz/go-craft/agent"
)

const reach = 4.5

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	bot, err := agent.Join(ctx, "localhost:25565", "gocraft_dig")
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
