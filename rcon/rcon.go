package rcon

import (
	"bufio"
	"net"
)

type Client struct {
	conn    net.Conn
	reader  *bufio.Reader
	counter int32
}

func Dial(address, password string) (*Client, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	client := &Client{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}
	if err := client.login(password); err != nil {
		_ = conn.Close()

		return nil, err
	}

	return client, nil
}

// a rejected password comes back as a frame carrying the sentinel request id
// instead of an error, which is the only way to tell the two apart
func (c *Client) login(password string) error {
	sent, err := c.send(frameLogin, password)
	if err != nil {
		return err
	}

	answer, err := readFrame(c.reader)
	if err != nil {
		return err
	}
	if answer.id == refusedID || answer.id != sent {
		return errRefused
	}

	return nil
}

func (c *Client) Run(command string) (string, error) {
	if _, err := c.send(frameCommand, command); err != nil {
		return "", err
	}

	answer, err := readFrame(c.reader)
	if err != nil {
		return "", err
	}

	return answer.payload, nil
}

func (c *Client) send(request frameType, payload string) (int32, error) {
	c.counter++

	outgoing := frame{
		id:      c.counter,
		typed:   request,
		payload: payload,
	}

	_, err := c.conn.Write(outgoing.encode())

	return c.counter, err
}

func (c *Client) Close() error {
	return c.conn.Close()
}
