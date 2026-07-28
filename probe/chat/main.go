package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/lrnxzz/go-craft/agent"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	bot, err := agent.Join(ctx, "localhost:25565", "gocraft_chat")
	if err != nil {
		return err
	}

	echo := func(line string) {
		fmt.Printf("CHAT: %s\n", line)
	}
	bot.OnChat(echo)

	go func() {
		if err := bot.Run(ctx); err != nil {
			log.Println("run:", err)
		}
	}()

	if err := bot.Ready(ctx); err != nil {
		return err
	}
	time.Sleep(time.Second)

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
