package host

import (
	"context"
	"errors"
	"fmt"
	"time"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/agent"
	v765 "github.com/lrnxzz/go-craft/codec/v765"
)

// Plugin is the code a user writes. It runs once the bot is connected and the
// terrain around it has arrived, and the session lasts exactly as long as it
// does: returning winds the connection down.
type Plugin func(context.Context, *agent.Agent) error

// Run owns the whole lifecycle — connect, wait for the world, hand over, shut
// down — so a plugin never repeats the bring-up.
//
// The plugin runs on the calling goroutine, not a new one. That is deliberate:
// a window belongs to the main OS thread and GLFW will not be moved off it, so
// anything that opens one must be able to do it from main.
func Run(ctx context.Context, address, username string, plugin Plugin) error {
	ctx, stop := context.WithCancel(ctx)
	defer stop()

	if err := speaks(ctx, address); err != nil {
		return err
	}

	bot, err := agent.Join(ctx, address, username)
	if err != nil {
		return err
	}

	disconnected := make(chan error, 1)
	go func() {
		disconnected <- bot.Run(ctx)
	}()

	played := play(ctx, stop, bot, plugin)

	// a session that died on its own is the real cause; the plugin only ever saw
	// the context being cancelled out from under it
	dropped := <-disconnected
	if dropped != nil && !ended(dropped) {
		return dropped
	}

	return played
}

// a context that ran out is how a session is meant to end, not how it fails
func ended(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

func play(ctx context.Context, stop context.CancelFunc, bot *agent.Agent, plugin Plugin) error {
	defer stop()

	if err := bot.Ready(ctx); err != nil {
		return err
	}

	return plugin(ctx, bot)
}

// speaks asks the server which protocol it talks before anything tries to log
// in. A mismatched login gets no reply at all — the connection simply sits there
// — so without this the client waits for a handshake that will never come, with
// nothing on screen to say why.
//
// The check is a courtesy, not a gate: a server that refuses to be pinged is
// still worth trying, and only a version it states outright is worth refusing.
func speaks(ctx context.Context, address string) error {
	asking, stop := context.WithTimeout(ctx, statusTimeout)
	defer stop()

	status, err := gocraft.Ping(asking, address)
	if err != nil {
		return nil
	}
	if status.Version.Protocol == v765.ProtocolVersion {
		return nil
	}

	return fmt.Errorf("host: %s speaks protocol %d (%s) and this client speaks %d (1.20.4)",
		address, status.Version.Protocol, status.Version.Name, v765.ProtocolVersion)
}

const statusTimeout = 10 * time.Second
