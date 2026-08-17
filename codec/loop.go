package codec

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	packetBuffer = 256

	TickRate = 50 * time.Millisecond
)

var errClientClosed = errors.New("graft: client closed")

type ticker struct {
	rate time.Duration
	fn   func()
}

type loop struct {
	inbound  chan Packet
	handled  chan struct{}
	outbound chan Packet
	tick     ticker
	done     chan struct{}
	once     sync.Once
}

func (l *loop) send(packet Packet) error {
	select {
	case l.outbound <- packet:
		return nil
	case <-l.done:
		return errors.New("graft: send on a closed client")
	}
}

func (l *loop) close() {
	l.once.Do(func() {
		slog.Debug("disconnecting")
		close(l.done)
	})
}

func (l *loop) closed() bool {
	select {
	case <-l.done:
		return true
	default:
		return false
	}
}

func (c *Client) Tick(rate time.Duration, fn func()) {
	c.loop.tick = ticker{
		rate: rate,
		fn:   fn,
	}
}

func (c *Client) Run(parent context.Context) error {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	group, ctx := errgroup.WithContext(ctx)

	stop := context.AfterFunc(ctx, func() {
		_ = c.Close()
	})
	defer stop()

	group.Go(func() error {
		return c.receive(ctx)
	})
	group.Go(func() error {
		return c.dispatch(ctx)
	})
	group.Go(func() error {
		return c.transmit(ctx)
	})

	err := group.Wait()
	if errors.Is(err, errClientClosed) {
		return nil
	}

	return err
}

func (c *Client) receive(ctx context.Context) error {
	for {
		frame, err := c.transport.ReadFrame()
		if err != nil {
			return c.exit(ctx, err)
		}

		packet, known, err := c.protocol.Decode(c.State(), Clientbound, frame)
		if err != nil {
			return err
		}
		if !known {
			slog.Debug("skipped unknown packet", "id", frame.ID)

			continue
		}

		select {
		case c.loop.inbound <- packet:
		case <-ctx.Done():
			return ctx.Err()
		}

		// nothing is read until the handlers for this packet have run: one of
		// them may turn compression on or change the state, and both decide how
		// the very next frame is framed and decoded
		select {
		case <-c.loop.handled:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (c *Client) dispatch(ctx context.Context) error {
	var pulse <-chan time.Time
	if c.loop.tick.fn != nil {
		t := time.NewTicker(c.loop.tick.rate)
		defer t.Stop()
		pulse = t.C
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case packet := <-c.loop.inbound:
			slog.Debug("received", "packet", packet.Name())

			err := c.listeners.dispatch(c, packet)

			select {
			case c.loop.handled <- struct{}{}:
			case <-ctx.Done():
				return nil
			}

			if err != nil {
				return err
			}
		case <-pulse:
			c.loop.tick.fn()
		}
	}
}

func (c *Client) transmit(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case packet := <-c.loop.outbound:
			if err := c.transport.WriteFrame(EncodeFrame(packet)); err != nil {
				return err
			}

			slog.Debug("sent", "packet", packet.Name())
		}
	}
}

func (c *Client) exit(ctx context.Context, readErr error) error {
	switch {
	case ctx.Err() != nil:
		return ctx.Err()
	case c.loop.closed():
		return errClientClosed
	default:
		return readErr
	}
}
