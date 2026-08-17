package mojang

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/valyala/fasthttp"
)

const (
	loginWithXboxURL = "https://api.minecraftservices.com/authentication/login_with_xbox"
)

type identityRequest struct {
	IdentityToken string `json:"identityToken"`
}

type minecraftToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type Microsoft struct {
	ClientID string
	Prompt   func(DeviceCode)
	Flow     Flow
}

func (m Microsoft) Authenticate(ctx context.Context) (Session, error) {
	flow := m.Flow
	if flow.DeviceCodeURL == "" {
		flow = Live()
	}

	msa := NewMSA(m.ClientID)
	msa.Flow = flow

	slog.Debug("requesting device code")
	code, err := msa.RequestDeviceCode(ctx)
	if err != nil {
		return Session{}, err
	}

	if m.Prompt != nil {
		m.Prompt(code)
	}

	slog.Debug("awaiting authorization")
	tokens, err := msa.AwaitToken(ctx, code)
	if err != nil {
		return Session{}, err
	}

	xbox := NewXbox()
	xbox.Preamble = flow.XboxPreamble

	slog.Debug("authenticating with xbox")
	identity, err := xbox.Identity(ctx, tokens.AccessToken)
	if err != nil {
		return Session{}, err
	}

	slog.Debug("logging in to minecraft")
	minecraft, err := loginWithIdentity(ctx, identity)
	if err != nil {
		return Session{}, err
	}

	profile, err := NewMojang(minecraft.AccessToken).Profile(ctx)
	if err != nil {
		return Session{}, err
	}

	slog.Debug("logged in", "name", profile.Name)

	return Session{
		AccessToken: minecraft.AccessToken,
		Profile:     profile,
	}, nil
}

func loginWithIdentity(ctx context.Context, identityToken string) (minecraftToken, error) {
	body, err := json.Marshal(identityRequest{
		IdentityToken: identityToken,
	})
	if err != nil {
		return minecraftToken{}, err
	}

	call := request{
		method:      fasthttp.MethodPost,
		url:         loginWithXboxURL,
		contentType: jsonContentType,
		accept:      jsonContentType,
		body:        body,
		timeout:     timeout,
	}

	answer, status, err := send(ctx, call)
	if err != nil {
		return minecraftToken{}, err
	}
	if status != fasthttp.StatusOK {
		return minecraftToken{}, fmt.Errorf("mojang: login_with_xbox returned %d: %s", status, answer)
	}

	var token minecraftToken
	if err := json.Unmarshal(answer, &token); err != nil {
		return minecraftToken{}, err
	}

	return token, nil
}
