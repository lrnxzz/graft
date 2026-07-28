package main

import "github.com/spf13/cobra"

func genCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "gen",
		Short: "Generate codec sources and viewer assets",
	}

	command.AddCommand(blocksCommand())
	command.AddCommand(biomesCommand())
	command.AddCommand(itemsCommand())
	command.AddCommand(materialsCommand())
	command.AddCommand(atlasCommand())
	command.AddCommand(iconsCommand())

	return command
}
