package graft

import (
	"fmt"
	"net"
	"strconv"
)

// SplitAddress takes a host and port apart, defaulting the port when the address
// carries none. A bare host is the common way to write a server.
func SplitAddress(address string) (string, uint16, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return address, DefaultPort.Uint16(), nil
	}

	parsed, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return "", 0, fmt.Errorf("graft: invalid port %q in %q", port, address)
	}

	return host, uint16(parsed), nil
}
