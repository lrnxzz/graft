package mojang

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/valyala/fasthttp"
)

const (
	userAuthURL = "https://user.auth.xboxlive.com/user/authenticate"
	xstsAuthURL = "https://xsts.auth.xboxlive.com/xsts/authorize"

	userRelyingParty = "http://auth.xboxlive.com"
	xstsRelyingParty = "rp://api.minecraftservices.com/"

	rpsAuthMethod = "RPS"
	rpsSiteName   = "user.auth.xboxlive.com"
	retailSandbox = "RETAIL"
	jwtTokenType  = "JWT"
)

type xstsDenialCode uint64

const (
	xstsBanned              xstsDenialCode = 2148916227
	xstsNoXboxProfile       xstsDenialCode = 2148916233
	xstsRegionUnavailable   xstsDenialCode = 2148916235
	xstsAdultVerification   xstsDenialCode = 2148916236
	xstsAdultVerificationKR xstsDenialCode = 2148916237
	xstsChildAccount        xstsDenialCode = 2148916238
)

type XSTSError struct {
	XErr    xstsDenialCode `json:"XErr"`
	Message string         `json:"Message"`
}

func (e *XSTSError) Error() string {
	return fmt.Sprintf("mojang: xsts denied: %s (%d)", e.Reason(), e.XErr)
}

func (e *XSTSError) Reason() string {
	switch e.XErr {
	case xstsBanned:
		return "the account is banned from xbox"
	case xstsNoXboxProfile:
		return "the microsoft account has no xbox profile"
	case xstsRegionUnavailable:
		return "xbox live is unavailable in the account's region"
	case xstsAdultVerification, xstsAdultVerificationKR:
		return "the account needs adult verification"
	case xstsChildAccount:
		return "the account belongs to a minor and must be added to a family"
	}

	return "unknown denial"
}

type XboxToken struct {
	Token    string
	UserHash string
}

type xboxUser struct {
	UserHash string `json:"uhs"`
}

type xboxClaims struct {
	XUI []xboxUser `json:"xui"`
}

type xboxResponse struct {
	Token         string     `json:"Token"`
	DisplayClaims xboxClaims `json:"DisplayClaims"`
}

type userProperties struct {
	AuthMethod string `json:"AuthMethod"`
	SiteName   string `json:"SiteName"`
	RpsTicket  string `json:"RpsTicket"`
}

type userRequest struct {
	Properties   userProperties `json:"Properties"`
	RelyingParty string         `json:"RelyingParty"`
	TokenType    string         `json:"TokenType"`
}

type xstsProperties struct {
	SandboxID  string   `json:"SandboxId"`
	UserTokens []string `json:"UserTokens"`
}

type xstsRequest struct {
	Properties   xstsProperties `json:"Properties"`
	RelyingParty string         `json:"RelyingParty"`
	TokenType    string         `json:"TokenType"`
}

type Xbox struct {
	Client      *fasthttp.Client
	UserAuthURL string
	XSTSAuthURL string
	Preamble    string
}

func NewXbox() *Xbox {
	return &Xbox{
		Client:      &fasthttp.Client{},
		UserAuthURL: userAuthURL,
		XSTSAuthURL: xstsAuthURL,
		Preamble:    Live().XboxPreamble,
	}
}

func (x *Xbox) Identity(ctx context.Context, msaAccessToken string) (string, error) {
	user, err := x.UserToken(ctx, msaAccessToken)
	if err != nil {
		return "", err
	}

	session, err := x.XSTSToken(ctx, user)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("XBL3.0 x=%s;%s", session.UserHash, session.Token), nil
}

func (x *Xbox) UserToken(ctx context.Context, msaAccessToken string) (XboxToken, error) {
	payload, err := json.Marshal(userRequest{
		Properties: userProperties{
			AuthMethod: rpsAuthMethod,
			SiteName:   rpsSiteName,
			RpsTicket:  x.Preamble + msaAccessToken,
		},
		RelyingParty: userRelyingParty,
		TokenType:    jwtTokenType,
	})
	if err != nil {
		return XboxToken{}, err
	}

	target := x.UserAuthURL
	if target == "" {
		target = userAuthURL
	}

	return x.authorize(ctx, target, payload)
}

func (x *Xbox) XSTSToken(ctx context.Context, user XboxToken) (XboxToken, error) {
	payload, err := json.Marshal(xstsRequest{
		Properties: xstsProperties{
			SandboxID:  retailSandbox,
			UserTokens: []string{user.Token},
		},
		RelyingParty: xstsRelyingParty,
		TokenType:    jwtTokenType,
	})
	if err != nil {
		return XboxToken{}, err
	}

	target := x.XSTSAuthURL
	if target == "" {
		target = xstsAuthURL
	}

	return x.authorize(ctx, target, payload)
}

func (x *Xbox) authorize(ctx context.Context, target string, payload []byte) (XboxToken, error) {
	call := request{
		client:      x.Client,
		method:      fasthttp.MethodPost,
		url:         target,
		contentType: jsonContentType,
		accept:      jsonContentType,
		body:        payload,
		timeout:     timeout,
	}

	body, status, err := send(ctx, call)
	if err != nil {
		return XboxToken{}, err
	}

	if status == fasthttp.StatusUnauthorized {
		var denial XSTSError
		if err := json.Unmarshal(body, &denial); err == nil && denial.XErr != 0 {
			return XboxToken{}, &denial
		}
	}
	if status != fasthttp.StatusOK {
		return XboxToken{}, fmt.Errorf("mojang: xbox authorization returned %d: %s", status, body)
	}

	var decoded xboxResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return XboxToken{}, err
	}
	if len(decoded.DisplayClaims.XUI) == 0 {
		return XboxToken{}, errors.New("mojang: xbox response carries no user claims")
	}

	return XboxToken{
		Token:    decoded.Token,
		UserHash: decoded.DisplayClaims.XUI[0].UserHash,
	}, nil
}
