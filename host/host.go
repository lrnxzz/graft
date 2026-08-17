package host

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	graft "github.com/lrnxzz/graft"
	"github.com/lrnxzz/graft/agent"
	v765 "github.com/lrnxzz/graft/codec/v765"
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

	warnAboutProtocol(ctx, address)

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

	if err := spawn(ctx, bot); err != nil {
		return err
	}

	return plugin(ctx, bot)
}

// spawn waits for the world with a deadline, because a login that is accepted
// and then goes quiet is otherwise indistinguishable from a slow one. A proxy
// holding the bot in a waiting room looks exactly like this: logged in, ticking,
// and no terrain will ever arrive.
func spawn(ctx context.Context, bot *agent.Agent) error {
	waiting, stop := context.WithTimeout(ctx, spawnTimeout)
	defer stop()

	err := bot.Ready(waiting)
	if err == nil {
		return nil
	}
	if ctx.Err() != nil || !errors.Is(err, context.DeadlineExceeded) {
		return err
	}

	return fmt.Errorf("host: logged in but no world arrived in %s — the server may speak another "+
		"protocol, or be holding the bot somewhere before letting it in", spawnTimeout)
}

const spawnTimeout = 45 * time.Second

// warnAboutProtocol says what the server answers the status query with, when it
// is not what this client speaks. It only warns: a server running ViaVersion
// reports its own version here and still accepts an older client at login, so a
// refusal would turn a working connection into a refused one. If the login does
// then go quiet, this line is already in the log to explain it.
func warnAboutProtocol(ctx context.Context, address string) {
	asking, stop := context.WithTimeout(ctx, statusTimeout)
	defer stop()

	status, err := graft.Ping(asking, address)
	if err != nil || status.Version.Protocol == v765.ProtocolVersion {
		return
	}

	slog.Warn("the server answers with another protocol; trying anyway, since it may translate",
		"server", status.Version.Name,
		"speaks", status.Version.Protocol,
		"client", v765.ProtocolVersion)
}

const statusTimeout = 10 * time.Second
