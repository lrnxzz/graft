package v765_test

import (
	"context"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/lrnxzz/graft/codec"
	v765 "github.com/lrnxzz/graft/codec/v765"
)

func liveServer(t *testing.T) (host string, port uint16) {
	t.Helper()

	addr := os.Getenv("GRAFT_IT_ADDR")
	if addr == "" {
		t.Skip("set GRAFT_IT_ADDR to a running 1.20.4 server to run this integration test")
	}

	host, raw, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("GRAFT_IT_ADDR %q: %v", addr, err)
	}

	parsed, err := strconv.ParseUint(raw, 10, 16)
	if err != nil {
		t.Fatalf("GRAFT_IT_ADDR port %q: %v", raw, err)
	}

	return host, uint16(parsed)
}

func TestJoinReachesPlay(t *testing.T) {
	host, port := liveServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := codec.Dial(ctx, net.JoinHostPort(host, strconv.Itoa(int(port))))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}

	client := codec.NewClient(conn, v765.Protocol())

	joined := make(chan *v765.JoinGame, 1)
	ready := func(c *codec.Client, join *v765.JoinGame) error {
		joined <- join

		return c.Close()
	}

	if _, err := v765.Join(client, host, port, "graft_it", ready); err != nil {
		t.Fatalf("join: %v", err)
	}

	if err := client.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}

	select {
	case join := <-joined:
		if join.EntityID == 0 {
			t.Error("joined with entity id 0, want a server-assigned id")
		}
		t.Logf("reached play: entity=%d dimension=%s gamemode=%d", join.EntityID, join.DimensionName, join.GameMode)
	default:
		t.Fatal("client stopped before reaching play (JoinGame never fired)")
	}
}
