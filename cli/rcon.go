package main

import (
	"fmt"

	"github.com/lrnxzz/graft/rcon"
	"github.com/spf13/cobra"
)

func rconCommand() *cobra.Command {
	var password string

	command := &cobra.Command{
		Use:   "rcon <host[:port]> <command>...",
		Short: "Run console commands on a server over rcon",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			console, err := rcon.Dial(args[0], password)
			if err != nil {
				return err
			}
			defer func() {
				_ = console.Close()
			}()

			for _, line := range args[1:] {
				answer, err := console.Run(line)
				if err != nil {
					return err
				}

				fmt.Println(answer)
			}

			return nil
		},
	}

	command.Flags().StringVar(&password, "password", "graft", "rcon password")

	return command
}
