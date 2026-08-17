package main

import (
	"context"
	"fmt"
	"time"

	graft "github.com/lrnxzz/graft"
	"github.com/spf13/cobra"
)

func pingCommand() *cobra.Command {
	var timeout time.Duration

	command := &cobra.Command{
		Use:   "ping <host[:port]>",
		Short: "Query a server's status and latency",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			defer cancel()

			status, err := graft.Ping(ctx, args[0])
			if err != nil {
				return err
			}

			fmt.Println(status.MOTD())
			fmt.Printf("version  %s (protocol %d)\n", status.Version.Name, status.Version.Protocol)
			fmt.Printf("players  %d/%d\n", status.Players.Online, status.Players.Max)
			fmt.Printf("latency  %s\n", status.Latency.Round(time.Millisecond))

			return nil
		},
	}

	command.Flags().DurationVar(&timeout, "timeout", 5*time.Second, "status exchange deadline")

	return command
}
