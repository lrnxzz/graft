package gocraft

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	handshakeID      VarInt = 0x00
	statusRequestID  VarInt = 0x00
	statusResponseID VarInt = 0x00
	statusPingID     VarInt = 0x01

	stateStatus   VarInt = 1
	statusVersion VarInt = -1

	DefaultPort UShort = 25565
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

	conn, err := Dial(ctx, target)
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

	handshake := Frame{
		ID:      handshakeID,
		Payload: Marshal(statusVersion, host, port, stateStatus),
	}
	if err := conn.WriteFrame(handshake); err != nil {
		return Status{}, err
	}

	request := Frame{
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

	var body String
	if err := Unmarshal(response.Payload, &body); err != nil {
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

func measureLatency(conn *Conn) (time.Duration, error) {
	nonce := Long(time.Now().UnixMilli())
	start := time.Now()

	ping := Frame{
		ID:      statusPingID,
		Payload: Marshal(nonce),
	}
	if err := conn.WriteFrame(ping); err != nil {
		return 0, err
	}

	pong, err := conn.ReadFrame()
	if err != nil {
		return 0, err
	}

	latency := time.Since(start)

	var echoed Long
	if err := Unmarshal(pong.Payload, &echoed); err != nil {
		return 0, err
	}
	if pong.ID != statusPingID || echoed != nonce {
		return 0, fmt.Errorf("gocraft: pong 0x%02x with nonce %d does not match ping %d", pong.ID, echoed, nonce)
	}

	return latency, nil
}

func splitAddress(address string) (String, UShort, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return String(address), DefaultPort, nil
	}

	parsed, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return "", 0, fmt.Errorf("gocraft: invalid port %q in %q", port, address)
	}

	return String(host), UShort(parsed), nil
}
