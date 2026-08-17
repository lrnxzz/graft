package mojang_test

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/lrnxzz/graft/mojang"
)

func TestOfflineProfileIsDeterministic(t *testing.T) {
	provider := mojang.Offline{
		Username: "Bot01",
	}

	first, err := provider.Authenticate(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	second, err := provider.Authenticate(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if first.Profile != second.Profile {
		t.Errorf("offline profile is not deterministic: %+v vs %+v", first.Profile, second.Profile)
	}
	if first.Online() {
		t.Error("offline session reports an access token")
	}
	if first.Profile.Name != "Bot01" {
		t.Errorf("profile name = %q, want Bot01", first.Profile.Name)
	}
}

func TestOfflineUUIDLayout(t *testing.T) {
	usernames := []string{"Bot01", "Notch", "jeb_", "steve"}

	for _, username := range usernames {
		provider := mojang.Offline{
			Username: username,
		}

		session, err := provider.Authenticate(context.Background())
		if err != nil {
			t.Fatal(err)
		}

		raw, err := hex.DecodeString(session.Profile.ID)
		if err != nil {
			t.Errorf("profile id %q for %q is not hex: %v", session.Profile.ID, username, err)
			continue
		}
		if len(raw) != 16 {
			t.Errorf("profile id for %q decoded to %d bytes, want 16", username, len(raw))
			continue
		}

		version := raw[6] >> 4
		if version != 3 {
			t.Errorf("uuid version for %q = %d, want 3", username, version)
		}

		variant := raw[8] >> 6
		if variant != 0b10 {
			t.Errorf("uuid variant for %q = %#b, want 0b10", username, variant)
		}
	}
}

func TestOfflineRejectsEmptyUsername(t *testing.T) {
	provider := mojang.Offline{}

	if _, err := provider.Authenticate(context.Background()); err == nil {
		t.Error("expected an error, got nil")
	}
}
