package codec

import (
	"fmt"
	"log/slog"
	"sync/atomic"
)

type Transport interface {
	ReadFrame() (Frame, error)
	WriteFrame(Frame) error
	SetThreshold(int)
	Close() error
}

type Client struct {
	transport Transport
	protocol  *Protocol
	state     atomic.Uint32
	listeners listeners
	loop      loop
}

func NewClient(transport Transport, protocol *Protocol) *Client {
	return &Client{
		transport: transport,
		protocol:  protocol,
		listeners: make(listeners),
		loop: loop{
			inbound:  make(chan Packet),
			handled:  make(chan struct{}),
			outbound: make(chan Packet, packetBuffer),
			done:     make(chan struct{}),
		},
	}
}

func (c *Client) State() State {
	return State(c.state.Load())
}

func (c *Client) SetState(state State) {
	slog.Debug("switched state", "to", state)
	c.state.Store(uint32(state))
}

func (c *Client) Send(packet Packet) error {
	return c.loop.send(packet)
}

func (c *Client) SetCompression(threshold int) {
	slog.Debug("enabled compression", "threshold", threshold)
	c.transport.SetThreshold(threshold)
}

func (c *Client) Close() error {
	c.loop.close()

	return c.transport.Close()
}

func On[P Packet](c *Client, handler func(*Client, P) error) {
	var prototype P

	c.listeners.add(prototype.Name(), func(client *Client, packet Packet) error {
		typed, ok := packet.(P)
		if !ok {
			return fmt.Errorf("gocraft: handler for %s received %T", prototype.Name(), packet)
		}

		return handler(client, typed)
	})
}
