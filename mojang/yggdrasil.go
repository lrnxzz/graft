package mojang

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/valyala/fasthttp"
)

type yggdrasilAgent struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
}

type yggdrasilRequest struct {
	Agent       yggdrasilAgent `json:"agent"`
	Username    string         `json:"username"`
	Password    string         `json:"password"`
	ClientToken string         `json:"clientToken,omitempty"`
	RequestUser bool           `json:"requestUser"`
}

type yggdrasilProfile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type yggdrasilResponse struct {
	AccessToken     string           `json:"accessToken"`
	ClientToken     string           `json:"clientToken"`
	SelectedProfile yggdrasilProfile `json:"selectedProfile"`
}

type yggdrasilError struct {
	Err     string `json:"error"`
	Message string `json:"errorMessage"`
}

func (e *yggdrasilError) Error() string {
	return fmt.Sprintf("mojang: yggdrasil %s: %s", e.Err, e.Message)
}

type Yggdrasil struct {
	BaseURL     string
	Email       string
	Password    string
	ClientToken string
}

func (y Yggdrasil) Authenticate(ctx context.Context) (Session, error) {
	body, err := json.Marshal(yggdrasilRequest{
		Agent: yggdrasilAgent{
			Name:    "Minecraft",
			Version: 1,
		},
		Username:    y.Email,
		Password:    y.Password,
		ClientToken: y.ClientToken,
		RequestUser: false,
	})
	if err != nil {
		return Session{}, err
	}

	call := request{
		method:      fasthttp.MethodPost,
		url:         y.BaseURL + "/authenticate",
		contentType: jsonContentType,
		accept:      jsonContentType,
		body:        body,
		timeout:     timeout,
	}

	answer, status, err := send(ctx, call)
	if err != nil {
		return Session{}, err
	}

	if status != fasthttp.StatusOK {
		var failure yggdrasilError
		if err := json.Unmarshal(answer, &failure); err == nil && failure.Err != "" {
			return Session{}, &failure
		}

		return Session{}, fmt.Errorf("mojang: yggdrasil authenticate returned %d: %s", status, answer)
	}

	var decoded yggdrasilResponse
	if err := json.Unmarshal(answer, &decoded); err != nil {
		return Session{}, err
	}

	return Session{
		AccessToken: decoded.AccessToken,
		Profile: Profile{
			ID:   decoded.SelectedProfile.ID,
			Name: decoded.SelectedProfile.Name,
		},
	}, nil
}
