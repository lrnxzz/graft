package main

import (
	"context"
	"log/slog"
	"time"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/agent"
	"github.com/lrnxzz/go-craft/codec/v765/blocks"
	"github.com/lrnxzz/go-craft/host"
	"github.com/spf13/cobra"
)

func joinCommand() *cobra.Command {
	var (
		username string
		timeout  time.Duration
		walk     bool
	)

	command := &cobra.Command{
		Use:   "join <host[:port]>",
		Short: "Connect a bot to a server and simulate it until the timeout",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			defer cancel()

			slog.Info("connecting", "server", args[0], "as", username)

			observe := func(ctx context.Context, bot *agent.Agent) error {
				var start gocraft.Vec3d
				if walk {
					start = bot.Player().Position
					bot.Look(0, 0)
					bot.SetControls(gocraft.Controls{Forward: true})
					slog.Info("walking forward", "from", start)
				}

				<-ctx.Done()

				report(bot)
				if walk {
					end := bot.Player().Position
					slog.Info("walked", "to", end, "distance", end.Distance(start))
				}
				slog.Info("disconnected")

				return nil
			}

			return host.Run(ctx, args[0], username, observe)
		},
	}

	command.Flags().StringVar(&username, "username", "gocraft_bot", "bot username (offline mode)")
	command.Flags().DurationVar(&timeout, "timeout", 15*time.Second, "how long to stay connected")
	command.Flags().BoolVar(&walk, "walk", false, "walk forward after spawning")

	return command
}

func report(bot *agent.Agent) {
	player := bot.Player()
	feet := player.Position.Floor().Add(0, -1, 0)

	state, _ := bot.World().Block(feet.X, feet.Y, feet.Z)

	name := "air"
	if block, ok := blocks.Of(state); ok {
		name = string(block.Name)
	}

	slog.Info("perceived",
		"position", player.Position,
		"health", player.Health,
		"on_ground", player.OnGround,
		"loaded_chunks", bot.World().Loaded(),
		"standing_on", name)
}
