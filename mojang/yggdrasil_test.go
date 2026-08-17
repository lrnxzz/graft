package mojang_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/lrnxzz/graft/mojang"
)

func TestYggdrasilAuthenticate(t *testing.T) {
	baseURL := os.Getenv("GRAFT_YGGDRASIL_URL")
	email := os.Getenv("GRAFT_YGGDRASIL_EMAIL")
	password := os.Getenv("GRAFT_YGGDRASIL_PASSWORD")

	if baseURL == "" || email == "" || password == "" {
		t.Skip("GRAFT_YGGDRASIL_URL, GRAFT_YGGDRASIL_EMAIL and GRAFT_YGGDRASIL_PASSWORD not set")
	}

	provider := mojang.Yggdrasil{
		BaseURL:  baseURL,
		Email:    email,
		Password: password,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	session, err := provider.Authenticate(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if !session.Online() {
		t.Error("yggdrasil session carries no access token")
	}
	if session.Profile.Name == "" || len(session.Profile.ID) != 32 {
		t.Errorf("profile = %+v, want a name and 32-char id", session.Profile)
	}
}

func TestYggdrasilRejectsBadCredentials(t *testing.T) {
	baseURL := os.Getenv("GRAFT_YGGDRASIL_URL")
	if baseURL == "" {
		t.Skip("GRAFT_YGGDRASIL_URL not set")
	}

	provider := mojang.Yggdrasil{
		BaseURL:  baseURL,
		Email:    "does-not-exist@example.com",
		Password: "wrong-password",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if _, err := provider.Authenticate(ctx); err == nil {
		t.Error("expected an error, got nil")
	}
}
