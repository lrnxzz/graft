package gocraft

import (
	"github.com/lrnxzz/go-craft/codec"

	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	handshakeID      codec.VarInt = 0x00
	statusRequestID  codec.VarInt = 0x00
	statusResponseID codec.VarInt = 0x00
	statusPingID     codec.VarInt = 0x01

	stateStatus   codec.VarInt = 1
	statusVersion codec.VarInt = -1

	DefaultPort codec.UShort = 25565
)

type StatusVersion struct {
	Name     string `json:"name"`
	Protocol int    `json:"protocol"`
}

type StatusPlayer struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type StatusPlayers struct {
	Max    int            `json:"max"`
	Online int            `json:"online"`
	Sample []StatusPlayer `json:"sample"`
}

type Status struct {
	Version     StatusVersion   `json:"version"`
	Players     StatusPlayers   `json:"players"`
	Description json.RawMessage `json:"description"`
	Favicon     string          `json:"favicon"`
	Latency     time.Duration   `json:"-"`
}

type chatComponent struct {
	Text  string            `json:"text"`
	Extra []json.RawMessage `json:"extra"`
}

func (s Status) MOTD() string {
	return flattenChat(s.Description)
}

func flattenChat(raw json.RawMessage) string {
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return text
	}

	var component chatComponent
	if json.Unmarshal(raw, &component) != nil {
		return ""
	}

	var flattened strings.Builder

	flattened.WriteString(component.Text)
	for _, extra := range component.Extra {
		flattened.WriteString(flattenChat(extra))
	}

	return flattened.String()
}

func Ping(ctx context.Context, address string) (Status, error) {
	host, port, err := splitAddress(address)
	if err != nil {
		return Status{}, err
	}

	target := net.JoinHostPort(host.String(), strconv.Itoa(port.Int()))

	conn, err := codec.Dial(ctx, target)
	if err != nil {
		return Status{}, err
	}
	defer conn.Close()

	deadline, ok := ctx.Deadline()
	if ok {
		if err := conn.SetDeadline(deadline); err != nil {
			return Status{}, err
		}
	}

	handshake := codec.Frame{
		ID:      handshakeID,
		Payload: codec.Marshal(statusVersion, host, port, stateStatus),
	}
	if err := conn.WriteFrame(handshake); err != nil {
		return Status{}, err
	}

	request := codec.Frame{
		ID: statusRequestID,
	}
	if err := conn.WriteFrame(request); err != nil {
		return Status{}, err
	}

	response, err := conn.ReadFrame()
	if err != nil {
		return Status{}, err
	}
	if response.ID != statusResponseID {
		return Status{}, fmt.Errorf("gocraft: unexpected packet 0x%02x during status exchange", response.ID)
	}

	var body codec.String
	if err := codec.Unmarshal(response.Payload, &body); err != nil {
		return Status{}, err
	}

	var status Status
	if err := json.Unmarshal([]byte(body), &status); err != nil {
		return Status{}, fmt.Errorf("gocraft: malformed status response: %w", err)
	}

	latency, err := measureLatency(conn)
	if err != nil {
		return Status{}, err
	}

	status.Latency = latency

	return status, nil
}

func measureLatency(conn *codec.Conn) (time.Duration, error) {
	nonce := codec.Long(time.Now().UnixMilli())
	start := time.Now()

	ping := codec.Frame{
		ID:      statusPingID,
		Payload: codec.Marshal(nonce),
	}
	if err := conn.WriteFrame(ping); err != nil {
		return 0, err
	}

	pong, err := conn.ReadFrame()
	if err != nil {
		return 0, err
	}

	latency := time.Since(start)

	var echoed codec.Long
	if err := codec.Unmarshal(pong.Payload, &echoed); err != nil {
		return 0, err
	}
	if pong.ID != statusPingID || echoed != nonce {
		return 0, fmt.Errorf("gocraft: pong 0x%02x with nonce %d does not match ping %d", pong.ID, echoed, nonce)
	}

	return latency, nil
}

func splitAddress(address string) (codec.String, codec.UShort, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return codec.String(address), DefaultPort, nil
	}

	parsed, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return "", 0, fmt.Errorf("gocraft: invalid port %q in %q", port, address)
	}

	return codec.String(host), codec.UShort(parsed), nil
}
