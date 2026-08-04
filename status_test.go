package gocraft_test

import (
	"context"
	"encoding/json"
	"net"
	"testing"
	"time"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/codec"
)

func serveStatus(listener net.Listener, response codec.String) {
	transport, err := listener.Accept()
	if err != nil {
		return
	}
	defer transport.Close()

	server := codec.NewConn(transport)

	if _, err := server.ReadFrame(); err != nil {
		return
	}
	if _, err := server.ReadFrame(); err != nil {
		return
	}

	status := codec.Frame{
		ID:      0x00,
		Payload: codec.Marshal(response),
	}
	if err := server.WriteFrame(status); err != nil {
		return
	}

	ping, err := server.ReadFrame()
	if err != nil {
		return
	}

	pong := codec.Frame{
		ID:      0x01,
		Payload: ping.Payload,
	}
	server.WriteFrame(pong)
}

func TestPing(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	response := codec.String(`{
		"version": {"name": "1.20.4", "protocol": 765},
		"players": {"max": 20, "online": 3, "sample": [{"name": "steve", "id": "uuid"}]},
		"description": {"text": "go-craft ", "extra": [{"text": "test"}]}
	}`)

	go serveStatus(listener, response)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	status, err := gocraft.Ping(ctx, listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	if status.Version.Name != "1.20.4" || status.Version.Protocol != 765 {
		t.Errorf("version = %q protocol %d, want 1.20.4 protocol 765", status.Version.Name, status.Version.Protocol)
	}
	if status.Players.Online != 3 || status.Players.Max != 20 || len(status.Players.Sample) != 1 {
		t.Errorf("players = %d/%d with %d samples, want 3/20 with 1", status.Players.Online, status.Players.Max, len(status.Players.Sample))
	}
	if motd := status.MOTD(); motd != "go-craft test" {
		t.Errorf("MOTD() = %q, want %q", motd, "go-craft test")
	}
	if status.Latency <= 0 {
		t.Errorf("latency = %s, want > 0", status.Latency)
	}
}

func TestStatusMOTD(t *testing.T) {
	tests := []struct {
		description string
		want        string
	}{
		{
			description: `"plain motd"`,
			want:        "plain motd",
		},
		{
			description: `{"text": "a", "extra": [{"text": "b", "extra": [{"text": "c"}]}, "d"]}`,
			want:        "abcd",
		},
		{
			description: `[]`,
			want:        "",
		},
	}

	for _, tt := range tests {
		status := gocraft.Status{
			Description: json.RawMessage(tt.description),
		}

		if got := status.MOTD(); got != tt.want {
			t.Errorf("MOTD(%s) = %q, want %q", tt.description, got, tt.want)
		}
	}
}
