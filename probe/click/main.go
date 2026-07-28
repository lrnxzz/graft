package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/lrnxzz/go-craft/agent"
	"github.com/lrnxzz/go-craft/codec/v765/items"
	"github.com/lrnxzz/go-craft/host"
)

const (
	address    = "localhost:25565"
	targetSlot = 20
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	move := func(ctx context.Context, bot *agent.Agent) error {
		for bot.Inventory().Count(items.Bread) == 0 {
			time.Sleep(200 * time.Millisecond)
		}

		source, stocked := bot.Inventory().FindItem(items.Bread)
		if !stocked {
			return errors.New("probe: no bread found")
		}
		fmt.Printf("bread at slot %d, carried=%v\n", source, bot.Carried())

		if err := bot.ClickSlot(source); err != nil {
			return err
		}
		fmt.Printf("picked up: carried=%v slot=%v\n", bot.Carried(), bot.Inventory().Slot(source))

		if err := bot.ClickSlot(targetSlot); err != nil {
			return err
		}
		fmt.Printf("placed: carried=%v target=%v\n", bot.Carried(), bot.Inventory().Slot(targetSlot))

		time.Sleep(2 * time.Second)
		fmt.Printf("after server settle: target=%v source=%v carried=%v\n",
			bot.Inventory().Slot(targetSlot), bot.Inventory().Slot(source), bot.Carried())

		if bot.Inventory().Slot(targetSlot).Is(items.Bread) && bot.Carried().Empty() {
			fmt.Println("CLICK OK: server accepted the move")

			return nil
		}

		fmt.Println("CLICK FAILED: server reverted the move")

		return nil
	}

	return host.Run(ctx, address, "gocraft_click", move)
}
