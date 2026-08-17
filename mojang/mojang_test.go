package mojang_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"testing"
	"time"

	"github.com/lrnxzz/graft/mojang"
)

func sessionToken(t *testing.T) string {
	t.Helper()

	token := os.Getenv("GRAFT_ACCESS_TOKEN")
	if token == "" {
		t.Skip("GRAFT_ACCESS_TOKEN not set")
	}

	return token
}

func TestMojangProfile(t *testing.T) {
	client := mojang.NewMojang(sessionToken(t))

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	profile, err := client.Profile(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(profile.ID) != 32 {
		t.Errorf("profile id = %q, want 32 hex chars", profile.ID)
	}
	if profile.Name == "" {
		t.Error("profile name is empty")
	}
}

func TestMojangProfileWithInvalidToken(t *testing.T) {
	sessionToken(t)

	client := mojang.NewMojang("invalid-token")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if _, err := client.Profile(ctx); err == nil {
		t.Error("expected an error, got nil")
	}
}

func TestMojangJoinAndHasJoined(t *testing.T) {
	client := mojang.NewMojang(sessionToken(t))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	profile, err := client.Profile(ctx)
	if err != nil {
		t.Fatal(err)
	}

	nonce := make([]byte, 20)
	if _, err := rand.Read(nonce); err != nil {
		t.Fatal(err)
	}

	serverID := hex.EncodeToString(nonce)

	if err := client.JoinServer(ctx, profile, serverID); err != nil {
		t.Fatal(err)
	}

	joined, err := client.HasJoined(ctx, profile.Name, serverID)
	if err != nil {
		t.Fatal(err)
	}

	if joined.ID != profile.ID {
		t.Errorf("hasJoined returned id %q, want %q", joined.ID, profile.ID)
	}
}
