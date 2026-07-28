package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

// the long-running commands hang their deadline off this context; without a
// signal-aware one a ctrl-c would tear the connection down mid-packet
func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return root().ExecuteContext(ctx)
}

func root() *cobra.Command {
	var verbose bool

	command := &cobra.Command{
		Use:           "gocraft",
		Short:         "Minecraft protocol toolbox",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			level := slog.LevelInfo
			if verbose {
				level = slog.LevelDebug
			}

			options := slog.HandlerOptions{
				Level: level,
			}
			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &options)))
		},
	}

	command.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable debug logging")

	command.AddCommand(pingCommand())
	command.AddCommand(loginCommand())
	command.AddCommand(joinCommand())
	command.AddCommand(gotoCommand())
	command.AddCommand(genCommand())

	return command
}
