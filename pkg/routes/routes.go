// Package routes reads the system routing table (Linux via /proc/net/route).
package routes

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
)

// RTF bits from Linux include/uapi/linux/route.h (as shown in /proc/net/route).
const (
	rtfUp        = 0x0001
	rtfGateway   = 0x0002
	rtfHost      = 0x0004
	rtfReinstate = 0x0008
	rtfDynamic   = 0x0010
	rtfModified  = 0x0020
	rtfReject    = 0x0200
)

// Route is one row from the routing table.
type Route struct {
	Iface       string
	Destination string
	Gateway     string
	Flags       string
	RefCnt      string
	Use         string
	Metric      string
	Mask        string
}

// List returns kernel IPv4 routes. On Linux this reads /proc/net/route.
func List() ([]Route, error) {
	if runtime.GOOS != "linux" {
		return nil, fmt.Errorf("routes: only supported on linux (current GOOS=%s)", runtime.GOOS)
	}
	data, err := os.ReadFile("/proc/net/route")
	if err != nil {
		return nil, fmt.Errorf("routes: read /proc/net/route: %w", err)
	}
	return parseProcNetRoute(string(data))
}

func parseProcNetRoute(content string) ([]Route, error) {
	sc := bufio.NewScanner(strings.NewReader(content))
	if !sc.Scan() {
		return nil, fmt.Errorf("routes: empty /proc/net/route")
	}
	header := strings.Fields(sc.Text())
	if len(header) < 2 || header[0] != "Iface" {
		return nil, fmt.Errorf("routes: unexpected header in /proc/net/route")
	}
	var routes []Route
	lineNum := 1
	for sc.Scan() {
		lineNum++
		fields := strings.Fields(sc.Text())
		if len(fields) < 8 {
			continue
		}
		iface := fields[0]
		destHex := fields[1]
		gwHex := fields[2]
		flags := fields[3]
		refCnt := fields[4]
		use := fields[5]
		metric := fields[6]
		maskHex := fields[7]

		destIP, err := parseProcHexIPv4(destHex)
		if err != nil {
			return nil, fmt.Errorf("routes: line %d destination: %w", lineNum, err)
		}
		gwIP, err := parseProcHexIPv4(gwHex)
		if err != nil {
			return nil, fmt.Errorf("routes: line %d gateway: %w", lineNum, err)
		}
		maskIP, err := parseProcHexIPv4(maskHex)
		if err != nil {
			return nil, fmt.Errorf("routes: line %d mask: %w", lineNum, err)
		}

		routes = append(routes, Route{
			Iface:       iface,
			Destination: destIP.String(),
			Gateway:     gwIP.String(),
			Flags:       flags,
			RefCnt:      refCnt,
			Use:         use,
			Metric:      metric,
			Mask:        maskIP.String(),
		})
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return routes, nil
}

// FormatDestination returns a net-tools style destination: "default" for the
// default IPv4 route (0.0.0.0 / 0.0.0.0), otherwise the dotted quad.
func FormatDestination(dest, mask string) string {
	if dest == "0.0.0.0" && mask == "0.0.0.0" {
		return "default"
	}
	return dest
}

// FormatRTFFlags converts the hex RTF field from /proc/net/route into the
// flag letters printed by net-tools `route` (e.g. "UG", "U").
func FormatRTFFlags(flagsHex string) string {
	u, err := strconv.ParseUint(flagsHex, 16, 32)
	if err != nil {
		return flagsHex
	}
	var b strings.Builder
	if u&rtfUp != 0 {
		b.WriteByte('U')
	}
	if u&rtfGateway != 0 {
		b.WriteByte('G')
	}
	if u&rtfHost != 0 {
		b.WriteByte('H')
	}
	if u&rtfReinstate != 0 {
		b.WriteByte('R')
	}
	if u&rtfDynamic != 0 {
		b.WriteByte('D')
	}
	if u&rtfModified != 0 {
		b.WriteByte('M')
	}
	if u&rtfReject != 0 {
		b.WriteByte('!')
	}
	return b.String()
}

func parseProcHexIPv4(hexStr string) (net.IP, error) {
	if len(hexStr) != 8 {
		return nil, fmt.Errorf("invalid hex IPv4 field %q", hexStr)
	}
	u, err := strconv.ParseUint(hexStr, 16, 32)
	if err != nil {
		return nil, err
	}
	b := uint32(u)
	// /proc/net/route stores little-endian IPv4 in hex (e.g. 00000000 = 0.0.0.0).
	ip := net.IPv4(byte(b), byte(b>>8), byte(b>>16), byte(b>>24))
	return ip, nil
}
