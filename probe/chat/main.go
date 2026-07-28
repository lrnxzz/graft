package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/lrnxzz/go-craft/agent"
	"github.com/lrnxzz/go-craft/host"
)

const address = "localhost:25565"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	talk := func(ctx context.Context, bot *agent.Agent) error {
		agent.On(bot, func(e agent.ChatReceived) {
			fmt.Printf("CHAT: %s\n", e.Line)
		})

		if err := bot.Chat("hello from gocraft"); err != nil {
			return err
		}
		if err := bot.Chat("/say announcement through say"); err != nil {
			return err
		}

		time.Sleep(8 * time.Second)
		fmt.Println("done")

		return nil
	}

	return host.Run(ctx, address, "gocraft_chat", talk)
}
