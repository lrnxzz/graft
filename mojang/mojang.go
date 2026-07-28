package mojang

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/valyala/fasthttp"
)

const (
	profileURL   = "https://api.minecraftservices.com/minecraft/profile"
	joinURL      = "https://sessionserver.mojang.com/session/minecraft/join"
	hasJoinedURL = "https://sessionserver.mojang.com/session/minecraft/hasJoined"

	mojangTimeout = 10 * time.Second
)

type Profile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type joinRequest struct {
	AccessToken     string `json:"accessToken"`
	SelectedProfile string `json:"selectedProfile"`
	ServerID        string `json:"serverId"`
}

type Mojang struct {
	Client       *fasthttp.Client
	ProfileURL   string
	JoinURL      string
	HasJoinedURL string

	token string
}

func NewMojang(token string) *Mojang {
	return &Mojang{
		Client:       &fasthttp.Client{},
		ProfileURL:   profileURL,
		JoinURL:      joinURL,
		HasJoinedURL: hasJoinedURL,
		token:        token,
	}
}

func (m *Mojang) Profile(ctx context.Context) (Profile, error) {
	raw, status, err := m.do(ctx, fasthttp.MethodGet, m.ProfileURL, nil)
	if err != nil {
		return Profile{}, err
	}
	if status != fasthttp.StatusOK {
		return Profile{}, fmt.Errorf("mojang: profile request returned %d: %s", status, raw)
	}

	var profile Profile
	if err := json.Unmarshal(raw, &profile); err != nil {
		return Profile{}, err
	}

	return profile, nil
}

func (m *Mojang) JoinServer(ctx context.Context, profile Profile, serverID string) error {
	join := joinRequest{
		AccessToken:     m.token,
		SelectedProfile: profile.ID,
		ServerID:        serverID,
	}

	body, err := json.Marshal(join)
	if err != nil {
		return err
	}

	raw, status, err := m.do(ctx, fasthttp.MethodPost, m.JoinURL, body)
	if err != nil {
		return err
	}
	if status != fasthttp.StatusNoContent {
		return fmt.Errorf("mojang: join request returned %d: %s", status, raw)
	}

	return nil
}

func (m *Mojang) HasJoined(ctx context.Context, username, serverID string) (Profile, error) {
	target := fmt.Sprintf("%s?username=%s&serverId=%s", m.HasJoinedURL, url.QueryEscape(username), url.QueryEscape(serverID))

	raw, status, err := m.do(ctx, fasthttp.MethodGet, target, nil)
	if err != nil {
		return Profile{}, err
	}
	if status == fasthttp.StatusNoContent {
		return Profile{}, fmt.Errorf("mojang: no session for %s on server %s", username, serverID)
	}
	if status != fasthttp.StatusOK {
		return Profile{}, fmt.Errorf("mojang: hasJoined request returned %d: %s", status, raw)
	}

	var profile Profile
	if err := json.Unmarshal(raw, &profile); err != nil {
		return Profile{}, err
	}

	return profile, nil
}

func (m *Mojang) do(ctx context.Context, method, target string, body []byte) (raw []byte, status int, err error) {
	call := request{
		client:  m.Client,
		method:  method,
		url:     target,
		accept:  jsonContentType,
		bearer:  m.token,
		body:    body,
		timeout: mojangTimeout,
	}
	if body != nil {
		call.contentType = jsonContentType
	}

	return send(ctx, call)
}
