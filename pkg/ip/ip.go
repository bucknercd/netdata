// Package ip discovers public and per-interface IP addresses.
package ip

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// InterfaceAddrs holds the unicast addresses assigned to one interface.
type InterfaceAddrs struct {
	Name  string
	Addrs []string
}

// Public fetches this host's public IPv4/IPv6 address seen on the internet
// using an HTTPS request (default: api.ipify.org). Requires outbound HTTPS.
func Public() (string, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	return fetchPublic(client, "https://api.ipify.org")
}

func fetchPublic(client *http.Client, url string) (string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("public ip: request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("public ip: HTTP %s", resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 256))
	if err != nil {
		return "", fmt.Errorf("public ip: read body: %w", err)
	}
	s := strings.TrimSpace(string(body))
	if s == "" {
		return "", fmt.Errorf("public ip: empty response")
	}
	parsed := net.ParseIP(s)
	if parsed == nil {
		return "", fmt.Errorf("public ip: invalid address %q", s)
	}
	return parsed.String(), nil
}

// ListCurrent returns usable unicast IP addresses per interface (system order).
// Link-local and unspecified addresses are skipped; loopback is included on lo.
// Interfaces with no matching addresses are omitted.
func ListCurrent() ([]InterfaceAddrs, error) {
	raw, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var out []InterfaceAddrs
	for _, ifi := range raw {
		addrs, aerr := ifi.Addrs()
		if aerr != nil {
			return nil, fmt.Errorf("%s: %w", ifi.Name, aerr)
		}
		var ips []string
		for _, a := range addrs {
			var ip net.IP
			switch v := a.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			default:
				continue
			}
			if ip == nil || ip.IsUnspecified() || ip.IsLinkLocalUnicast() {
				continue
			}
			if ip4 := ip.To4(); ip4 != nil {
				ip = ip4
			}
			ips = appendUniqueString(ips, ip.String())
		}
		if len(ips) == 0 {
			continue
		}
		out = append(out, InterfaceAddrs{Name: ifi.Name, Addrs: ips})
	}
	return out, nil
}

func appendUniqueString(ips []string, s string) []string {
	for _, x := range ips {
		if x == s {
			return ips
		}
	}
	return append(ips, s)
}
