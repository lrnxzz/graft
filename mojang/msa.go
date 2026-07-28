package mojang

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/valyala/fasthttp"
)

const (
	formContentType = "application/x-www-form-urlencoded"

	msaTimeout = 10 * time.Second
)

type MSA struct {
	Client   *fasthttp.Client
	ClientID string
	Flow     Flow
}

func NewMSA(clientID string) *MSA {
	return &MSA{
		Client:   &fasthttp.Client{},
		ClientID: clientID,
		Flow:     Live(),
	}
}

func (m *MSA) RequestDeviceCode(ctx context.Context) (DeviceCode, error) {
	form := msaForm{
		clientID:     m.ClientID,
		scope:        m.Flow.Scope,
		responseType: m.Flow.ResponseType,
	}

	call := request{
		client:      m.Client,
		method:      fasthttp.MethodPost,
		url:         m.Flow.DeviceCodeURL,
		contentType: formContentType,
		body:        []byte(form.encode()),
		timeout:     msaTimeout,
	}

	body, status, err := send(ctx, call)
	if err != nil {
		return DeviceCode{}, err
	}
	if status != fasthttp.StatusOK {
		return DeviceCode{}, fmt.Errorf("mojang: device code request returned %d: %s", status, body)
	}

	var code DeviceCode
	if err := json.Unmarshal(body, &code); err != nil {
		return DeviceCode{}, err
	}

	return code, nil
}

func (m *MSA) AwaitToken(ctx context.Context, code DeviceCode) (TokenSet, error) {
	poll := devicePoll{
		msa:      m,
		code:     code,
		interval: code.PollInterval(),
		expiry:   time.Now().Add(code.Lifetime()),
	}

	return poll.run(ctx)
}

func (m *MSA) Refresh(ctx context.Context, refreshToken string) (TokenSet, error) {
	form := msaForm{
		clientID:     m.ClientID,
		grant:        grantRefresh,
		refreshToken: refreshToken,
		scope:        m.Flow.Scope,
	}

	return m.token(ctx, form)
}

func (m *MSA) token(ctx context.Context, form msaForm) (TokenSet, error) {
	call := request{
		client:      m.Client,
		method:      fasthttp.MethodPost,
		url:         m.Flow.TokenURL,
		contentType: formContentType,
		body:        []byte(form.encode()),
		timeout:     msaTimeout,
	}

	body, status, err := send(ctx, call)
	if err != nil {
		return TokenSet{}, err
	}

	if status == fasthttp.StatusOK {
		var tokens TokenSet
		if err := json.Unmarshal(body, &tokens); err != nil {
			return TokenSet{}, err
		}

		return tokens, nil
	}

	var failure OAuthError
	if err := json.Unmarshal(body, &failure); err != nil || failure.Code == "" {
		return TokenSet{}, fmt.Errorf("mojang: token request returned %d: %s", status, body)
	}

	return TokenSet{}, &failure
}

type devicePoll struct {
	msa      *MSA
	code     DeviceCode
	interval time.Duration
	expiry   time.Time
}

func (p *devicePoll) run(ctx context.Context) (TokenSet, error) {
	for time.Now().Before(p.expiry) {
		if err := p.wait(ctx); err != nil {
			return TokenSet{}, err
		}

		tokens, authorized, err := p.attempt(ctx)
		if err != nil {
			return TokenSet{}, err
		}
		if authorized {
			return tokens, nil
		}
	}

	return TokenSet{}, errors.New("mojang: device code expired before authorization")
}

func (p *devicePoll) wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(p.interval):
		return nil
	}
}

func (p *devicePoll) attempt(ctx context.Context) (TokenSet, bool, error) {
	form := msaForm{
		clientID:   p.msa.ClientID,
		grant:      grantDeviceCode,
		deviceCode: p.code.DeviceCode,
	}

	tokens, err := p.msa.token(ctx, form)
	if err == nil {
		return tokens, true, nil
	}

	var failure *OAuthError
	if !errors.As(err, &failure) {
		return TokenSet{}, false, err
	}

	switch failure.Code {
	case oauthAuthorizationPending:
		return TokenSet{}, false, nil
	case oauthSlowDown:
		p.interval += slowDownPenalty
		return TokenSet{}, false, nil
	}

	return TokenSet{}, false, failure
}
