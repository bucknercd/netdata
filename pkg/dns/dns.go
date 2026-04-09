// Package dns reads system DNS resolver configuration (resolv.conf format).
package dns

import (
	"fmt"
	"os"
	"strings"
)

const (
	defaultResolvPath         = "/etc/resolv.conf"
	systemdResolvedResolvPath = "/run/systemd/resolve/resolv.conf"
)

// Info holds DNS client configuration from a resolv.conf file.
type Info struct {
	Nameservers []string
	Search      []string
	Domain      string
	Options     []string
	// RecursiveResolvers are upstream nameservers systemd-resolved forwards to,
	// when they differ from Nameservers (e.g. libc sees 127.0.0.53 only).
	RecursiveResolvers []string
}

// Load reads and parses the file at path. See [LoadSystem] for /etc/resolv.conf.
func Load(path string) (Info, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Info{}, fmt.Errorf("dns: read %s: %w", path, err)
	}
	return ParseResolvConf(string(data))
}

// LoadSystem parses /etc/resolv.conf and, when available, merges upstream
// nameservers from systemd-resolved's resolv.conf (real recursive resolvers).
func LoadSystem() (Info, error) {
	info, err := Load(defaultResolvPath)
	if err != nil {
		return Info{}, err
	}
	data, rerr := os.ReadFile(systemdResolvedResolvPath)
	if rerr != nil {
		return info, nil
	}
	sys, perr := ParseResolvConf(string(data))
	if perr != nil || len(sys.Nameservers) == 0 {
		return info, nil
	}
	if !stringSliceEqualSet(info.Nameservers, sys.Nameservers) {
		info.RecursiveResolvers = append([]string(nil), sys.Nameservers...)
	}
	return info, nil
}

func stringSliceEqualSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	count := make(map[string]int, len(a))
	for _, s := range a {
		count[s]++
	}
	for _, s := range b {
		if count[s] == 0 {
			return false
		}
		count[s]--
	}
	return true
}

// ParseResolvConf parses resolv.conf text (nameserver, search, domain, options).
func ParseResolvConf(content string) (Info, error) {
	var info Info
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		key := strings.ToLower(fields[0])
		switch key {
		case "nameserver":
			if len(fields) < 2 {
				return Info{}, fmt.Errorf("dns: invalid nameserver line: %q", line)
			}
			for _, ns := range fields[1:] {
				info.Nameservers = append(info.Nameservers, ns)
			}
		case "search":
			info.Search = append(info.Search, fields[1:]...)
		case "domain":
			if len(fields) < 2 {
				return Info{}, fmt.Errorf("dns: invalid domain line: %q", line)
			}
			if info.Domain == "" {
				info.Domain = fields[1]
			}
		case "options":
			if len(fields) < 2 {
				continue
			}
			info.Options = append(info.Options, fields[1:]...)
		default:
			// ignore unknown directives (e.g. sortlist, ndots in some parsers — keep minimal)
		}
	}
	return info, nil
}
