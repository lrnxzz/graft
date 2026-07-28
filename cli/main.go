package main

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var verbose bool

	root := &cobra.Command{
		Use:           "gocraft",
		Short:         "Minecraft protocol toolbox",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			level := slog.LevelInfo
			if verbose {
				level = slog.LevelDebug
			}

			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: level,
			})))
		},
	}

	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable debug logging")

	root.AddCommand(pingCommand())
	root.AddCommand(loginCommand())
	root.AddCommand(joinCommand())
	root.AddCommand(gotoCommand())
	root.AddCommand(genCommand())

	if err := root.Execute(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
