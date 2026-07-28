package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"

	"github.com/lrnxzz/go-craft/mojang"
	"github.com/spf13/cobra"
)

// the session carries a refresh token, so it lands in a file only its owner can
// read; stdout would leave a copy in shell history, a pipe or a CI log
const sessionMode = 0o600

var errNoClientID = errors.New("gocraft: a Microsoft client id is required (--client-id or GOCRAFT_CLIENT_ID)")

func loginCommand() *cobra.Command {
	var (
		clientID string
		output   string
	)

	command := &cobra.Command{
		Use:   "login",
		Short: "Authenticate a Microsoft account for online-mode servers",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if clientID == "" {
				return errNoClientID
			}

			announce := func(code mojang.DeviceCode) {
				slog.Info("authorize", "url", code.VerificationURI, "code", code.UserCode)
			}
			provider := mojang.Microsoft{
				ClientID: clientID,
				Prompt:   announce,
			}

			session, err := provider.Authenticate(cmd.Context())
			if err != nil {
				return err
			}

			encoded, err := json.MarshalIndent(session, "", "  ")
			if err != nil {
				return err
			}
			if err := os.WriteFile(output, encoded, sessionMode); err != nil {
				return err
			}

			slog.Info("logged in",
				"name", session.Profile.Name,
				"id", session.Profile.ID,
				"session", output)

			return nil
		},
	}

	command.Flags().StringVar(&clientID, "client-id", os.Getenv("GOCRAFT_CLIENT_ID"), "Microsoft application (Azure) client id")
	command.Flags().StringVar(&output, "out", "gocraft-session.json", "file the session is written to, readable only by you")

	return command
}
