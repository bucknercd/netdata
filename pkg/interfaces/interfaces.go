// Package interfaces lists local network interfaces using the standard library.
package interfaces

import (
	"fmt"
	"net"
)

// Interface describes one network interface and its addresses.
type Interface struct {
	Index   int
	Name    string
	MTU     int
	Flags   net.Flags
	HWAddr  string
	Addrs   []string
}

// List returns all system network interfaces with their addresses.
func List() ([]Interface, error) {
	raw, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	out := make([]Interface, 0, len(raw))
	for _, ifi := range raw {
		addrs, aerr := ifi.Addrs()
		if aerr != nil {
			return nil, fmt.Errorf("%s: %w", ifi.Name, aerr)
		}
		addrStrs := make([]string, len(addrs))
		for i, a := range addrs {
			addrStrs[i] = a.String()
		}
		hw := ""
		if len(ifi.HardwareAddr) > 0 {
			hw = ifi.HardwareAddr.String()
		}
		out = append(out, Interface{
			Index:  ifi.Index,
			Name:   ifi.Name,
			MTU:    ifi.MTU,
			Flags:  ifi.Flags,
			HWAddr: hw,
			Addrs:  addrStrs,
		})
	}
	return out, nil
}
