package main

import (
	"context"
	"log/slog"
	"time"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/agent"
	"github.com/lrnxzz/go-craft/pathfinder"
	"github.com/spf13/cobra"
)

func gotoCommand() *cobra.Command {
	var (
		username string
		goal     string
		timeout  time.Duration
	)

	command := &cobra.Command{
		Use:   "goto <host[:port]>",
		Short: "Join a server and walk the bot to a goal through the pathfinder",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target, err := gocraft.ParsePosition(goal)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			defer cancel()

			bot, err := agent.Join(ctx, args[0], username)
			if err != nil {
				return err
			}

			finished := make(chan error, 1)
			go func() {
				finished <- bot.Run(ctx)
			}()

			select {
			case <-bot.Spawned():
			case err := <-finished:
				return err
			}

			if err := bot.Ready(ctx); err != nil {
				return err
			}

			slog.Info("navigating", "from", bot.Snapshot().Position, "to", target)

			walked := make(chan error, 1)
			go func() {
				arrived, err := bot.Navigate(ctx, pathfinder.GoalAt(target))
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
					slog.Info("walking", "at", bot.Snapshot().Position)
				case err := <-walked:
					return err
				case err := <-finished:
					return err
				}
			}
		},
	}

	command.Flags().StringVar(&username, "username", "gocraft_pathfinder", "bot username")
	command.Flags().StringVar(&goal, "goal", "", "target block as x,y,z")
	command.Flags().DurationVar(&timeout, "timeout", 2*time.Minute, "abort after this long")

	return command
}
