package main

import (
	"context"
	"log/slog"
	"time"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/agent"
	"github.com/lrnxzz/go-craft/host"
	"github.com/lrnxzz/go-craft/pathfinder"
	"github.com/spf13/cobra"
)

func gotoCommand() *cobra.Command {
	var (
		username string
		timeout  time.Duration
	)

	command := &cobra.Command{
		Use:   "goto <host[:port]> <x,y,z>...",
		Short: "Join a server and walk the bot through one goal after another",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			legs, err := positions(args[1:])
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			defer cancel()

			march := func(ctx context.Context, bot *agent.Agent) error {
				for _, leg := range legs {
					if err := walk(ctx, bot, leg); err != nil {
						return err
					}
				}
				slog.Info("all legs complete")

				return nil
			}

			return host.Run(ctx, args[0], username, march)
		},
	}

	command.Flags().StringVar(&username, "username", "gocraft_pathfinder", "bot username")
	command.Flags().DurationVar(&timeout, "timeout", 2*time.Minute, "abort after this long")

	return command
}

func positions(written []string) ([]gocraft.Position, error) {
	legs := make([]gocraft.Position, 0, len(written))
	for _, each := range written {
		leg, err := gocraft.ParsePosition(each)
		if err != nil {
			return nil, err
		}

		legs = append(legs, leg)
	}

	return legs, nil
}

// walk reports where the bot is once a second while the leg runs. A planner that
// has stalled looks exactly like one still working without it, which is the
// whole reason to watch a walk rather than only read how it ended.
func walk(ctx context.Context, bot *agent.Agent, leg gocraft.Position) error {
	slog.Info("navigating", "from", bot.Snapshot().Position, "to", leg)

	walked := make(chan error, 1)
	go func() {
		arrived, err := bot.Navigate(ctx, pathfinder.GoalAt(leg))
		if err == nil {
			slog.Info("arrived", "at", arrived)
		}

		walked <- err
	}()

	progress := time.NewTicker(time.Second)
	defer progress.Stop()

	for {
		select {
		case <-progress.C:
			snapshot := bot.Snapshot()
			slog.Info("walking", "at", snapshot.Position, "ground", snapshot.OnGround)
		case err := <-walked:
			return err
		}
	}
}
