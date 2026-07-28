package main

import (
	"strings"

	"github.com/spf13/cobra"
)

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

func ident(name string) string {
	var b strings.Builder
	for _, part := range strings.Split(name, "_") {
		if part != "" {
			b.WriteString(strings.ToUpper(part[:1]) + part[1:])
		}
	}

	return b.String()
}
