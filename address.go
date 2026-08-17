package graft

import (
	"fmt"
	"net"
	"strconv"
	"strings"
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

// Resolve turns what a player would type into the host and port to dial. A bare
// host gets the SRV lookup the vanilla client does — most public servers hide
// the game host behind one, and dialing the written name would only reach their
// website. A written port pins the address and skips the redirect, also like
// vanilla.
func Resolve(address string) (string, uint16, error) {
	host, port, err := SplitAddress(address)
	if err != nil || strings.Contains(address, ":") {
		return host, port, err
	}

	_, records, err := net.LookupSRV("minecraft", "tcp", host)
	if err != nil || len(records) == 0 {
		return host, port, nil
	}

	return strings.TrimSuffix(records[0].Target, "."), records[0].Port, nil
}
