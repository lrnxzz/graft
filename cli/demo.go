package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/lrnxzz/go-craft/agent"
	"github.com/lrnxzz/go-craft/host"
	"github.com/lrnxzz/go-craft/pathfinder"
	"github.com/spf13/cobra"
)

const retryDelay = 3 * time.Second

func demoCommand() *cobra.Command {
	var username string

	command := &cobra.Command{
		Use:   "demo <host[:port]>",
		Short: "Walk the obstacle course carved by gocraft course, lap after lap",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return laps(cmd.Context(), args[0], username)
		},
	}

	command.Flags().StringVar(&username, "username", "gocraft_bot", "bot username")

	return command
}

// a demo outlives the sessions it is made of: the point is a bot that is still
// walking when someone looks over, so a dropped connection reconnects rather
// than ending the run
func laps(ctx context.Context, address, username string) error {
	slog.Info("demo starting", "server", address)

	for ctx.Err() == nil {
		if err := session(ctx, address, username); err != nil {
			slog.Warn("session ended", "err", err)
		}

		slog.Info("reconnecting", "in", retryDelay)

		select {
		case <-time.After(retryDelay):
		case <-ctx.Done():
		}
	}

	return nil
}

func session(ctx context.Context, address, username string) error {
	course := lap()

	walk := func(ctx context.Context, bot *agent.Agent) error {
		for round := 1; ; round++ {
			slog.Info("lap", "number", round)

			for _, goal := range course {
				slog.Info("heading out", "to", goal, "from", bot.Snapshot().Position)

				arrived, err := bot.Navigate(ctx, pathfinder.GoalAt(goal))
				if err != nil {
					slog.Warn("leg failed", "err", err)
				} else {
					slog.Info("arrived", "at", arrived)
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

	return host.Run(ctx, address, username, walk)
}
