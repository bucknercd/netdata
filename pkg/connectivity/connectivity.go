// Package connectivity runs short outbound probes to separate routing, DNS, and HTTPS issues.
package connectivity

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"forge/tmp/nettest/pkg/ip"
	"forge/tmp/nettest/pkg/routes"
)

const (
	// Overall wall-clock budget for all probes when run together.
	overallTimeout = 5 * time.Second
	tcpProbeAddr   = "1.1.1.1:443"
	dnsProbeName   = "one.one.one.one"
	// icmpPublicIPv4 is probed with ICMP echo (no DNS); matches the host in tcpProbeAddr.
	icmpPublicIPv4 = "1.1.1.1"
)

const noIPv4DefaultRouteDetail = "no IPv4 default route"

// gatewayTCPPorts are tried in order; connection refused still proves L3/L4 reachability.
var gatewayTCPPorts = []int{80, 443, 53}

// Report holds the outcome of parallel probes.
type Report struct {
	Gateway     GatewayProbe
	GatewayICMP ICMPProbe
	PublicICMP  ICMPProbe
	TCPNoDNS    Probe
	DNS         Probe
	HTTPS       Probe
	// PublicIP is set when the HTTPS probe returned an address successfully.
	PublicIP string
}

// GatewayProbe checks TCP reachability to the primary IPv4 default gateway
// (common routers answer on at least one of 80/443/53). See also [Report.GatewayICMP].
type GatewayProbe struct {
	Attempted bool
	OK        bool
	Detail    string
	Address   string
	Iface     string
}

// ICMPProbe is an IPv4 echo request to a single address (unprivileged udp4 when allowed, else raw).
type ICMPProbe struct {
	Attempted bool
	Target    string
	OK        bool
	Detail    string
	// NoSocket is true when the process could not open udp4 or raw ICMP (permissions / policy).
	NoSocket bool
}

// Probe is one check (TCP without DNS, DNS lookup, or HTTPS public IP).
type Probe struct {
	OK     bool
	Detail string
}

// Run executes TCP, DNS, and HTTPS probes in parallel. Each probe respects ctx;
// use [context.WithTimeout] so the tool never blocks for long when offline.
func Run(ctx context.Context) Report {
	ctx, cancel := context.WithTimeout(ctx, overallTimeout)
	defer cancel()

	var rep Report
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(6)

	go func() {
		defer wg.Done()
		g := probeDefaultGateway(ctx)
		mu.Lock()
		rep.Gateway = g
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		p := probeDefaultGatewayICMP(ctx)
		mu.Lock()
		rep.GatewayICMP = p
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		p := icmpEcho(ctx, icmpPublicIPv4)
		mu.Lock()
		rep.PublicICMP = p
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		p := probeTCP(ctx, tcpProbeAddr)
		mu.Lock()
		rep.TCPNoDNS = p
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		p := probeDNS(ctx, dnsProbeName)
		mu.Lock()
		rep.DNS = p
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		p, addr := probeHTTPSPublic(ctx)
		mu.Lock()
		rep.HTTPS = p
		rep.PublicIP = addr
		mu.Unlock()
	}()

	wg.Wait()
	return rep
}

// OverallTimeout is the maximum time [Run] waits for all probes combined.
func OverallTimeout() time.Duration {
	return overallTimeout
}

// TCPProbeAddr is the host:port used for the routing probe (no DNS).
func TCPProbeAddr() string {
	return tcpProbeAddr
}

// DNSProbeName is the hostname used for the resolver probe.
func DNSProbeName() string {
	return dnsProbeName
}

// ICMPPublicAddr is the IPv4 address used for the internet ICMP echo probe.
func ICMPPublicAddr() string {
	return icmpPublicIPv4
}

func resolveDefaultGateway() (gw, iface, skip string) {
	dr, err := routes.PrimaryIPv4DefaultRoute()
	if err != nil {
		if errors.Is(err, routes.ErrNoIPv4DefaultRoute) {
			return "", "", noIPv4DefaultRouteDetail
		}
		return "", "", err.Error()
	}
	g := dr.Gateway
	if ip := net.ParseIP(g); ip == nil || ip.IsUnspecified() {
		return "", dr.Iface, "default route has no next-hop IP (0.0.0.0)"
	}
	return g, dr.Iface, ""
}

func probeDefaultGateway(ctx context.Context) GatewayProbe {
	gw, iface, skip := resolveDefaultGateway()
	if skip != "" {
		return GatewayProbe{Attempted: false, Detail: skip, Iface: iface}
	}
	return probeGatewayTCP(ctx, gw, iface)
}

func probeDefaultGatewayICMP(ctx context.Context) ICMPProbe {
	gw, iface, skip := resolveDefaultGateway()
	if skip != "" {
		detail := skip
		if iface != "" && strings.Contains(skip, "next-hop") {
			detail = fmt.Sprintf("%s [iface %s]", skip, iface)
		}
		return ICMPProbe{Attempted: false, Detail: detail}
	}
	return icmpEcho(ctx, gw)
}

func probeGatewayTCP(ctx context.Context, gw, iface string) GatewayProbe {
	var last string
	for _, port := range gatewayTCPPorts {
		if err := ctx.Err(); err != nil {
			return GatewayProbe{
				Attempted: true,
				OK:        false,
				Detail:    err.Error(),
				Address:   gw,
				Iface:     iface,
			}
		}
		addr := net.JoinHostPort(gw, strconv.Itoa(port))
		d := net.Dialer{Timeout: 1200 * time.Millisecond}
		conn, err := d.DialContext(ctx, "tcp", addr)
		if err == nil {
			_ = conn.Close()
			return GatewayProbe{
				Attempted: true,
				OK:        true,
				Detail:    fmt.Sprintf("TCP %d connected", port),
				Address:   gw,
				Iface:     iface,
			}
		}
		if errors.Is(err, syscall.ECONNREFUSED) {
			return GatewayProbe{
				Attempted: true,
				OK:        true,
				Detail:    fmt.Sprintf("reachable (TCP %d: connection refused)", port),
				Address:   gw,
				Iface:     iface,
			}
		}
		last = err.Error()
	}
	return GatewayProbe{
		Attempted: true,
		OK:        false,
		Detail:    last,
		Address:   gw,
		Iface:     iface,
	}
}

func probeTCP(ctx context.Context, address string) Probe {
	d := net.Dialer{Timeout: 3 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", address)
	if err != nil {
		return Probe{OK: false, Detail: err.Error()}
	}
	_ = conn.Close()
	return Probe{OK: true, Detail: "connected"}
}

func probeDNS(ctx context.Context, name string) Probe {
	r := &net.Resolver{PreferGo: true}
	addrs, err := r.LookupHost(ctx, name)
	if err != nil {
		return Probe{OK: false, Detail: err.Error()}
	}
	if len(addrs) == 0 {
		return Probe{OK: false, Detail: "no addresses returned"}
	}
	return Probe{OK: true, Detail: strings.Join(addrs, ", ")}
}

func probeHTTPSPublic(ctx context.Context) (Probe, string) {
	addr, err := ip.PublicWithContext(ctx)
	if err != nil {
		return Probe{OK: false, Detail: err.Error()}, ""
	}
	return Probe{OK: true, Detail: "ok"}, addr
}

// Summary suggests the most likely failure layer from probe results.
func Summary(r Report) string {
	if !r.Gateway.Attempted && r.Gateway.Detail == noIPv4DefaultRouteDetail {
		return "No IPv4 default route in the kernel table — add a default gateway or bring up the interface that should provide it."
	}

	if r.Gateway.Attempted {
		if !r.Gateway.OK && !r.TCPNoDNS.OK {
			return fmt.Sprintf(
				"TCP to the default gateway %s (%s) did not succeed on ports 80/443/53, and the internet probe failed — check link to the gateway (Ethernet/Wi‑Fi), gateway power, and firewall.",
				r.Gateway.Address, r.Gateway.Iface,
			)
		}
		if !r.Gateway.OK && r.TCPNoDNS.OK {
			return "A public address was reachable, but TCP to the default gateway on 80/443/53 did not connect or refuse — the gateway may have no listener on those ports (often fine)."
		}
	}

	switch {
	case !r.TCPNoDNS.OK:
		return fmt.Sprintf("Likely routing or local link issue: TCP to %s (no DNS) failed. Fix default route / link / firewall before DNS or public IP will work.", tcpProbeAddr)
	case !r.DNS.OK:
		return "TCP to a public address works, but DNS lookup failed — check resolv.conf, systemd-resolved, VPN DNS, or captive portal."
	case !r.HTTPS.OK:
		return "TCP and DNS work, but HTTPS to the public-IP service failed — possible proxy, TLS interception, or outbound HTTPS filter."
	default:
		msg := "Outbound path, DNS, and HTTPS to the public-IP service all succeeded."
		if r.PublicICMP.Attempted && !r.PublicICMP.OK && !r.PublicICMP.NoSocket && r.TCPNoDNS.OK {
			msg += fmt.Sprintf(" Note: ICMP echo to %s failed while TCP to %s succeeded — ICMP is often filtered; that is normal.", icmpPublicIPv4, tcpProbeAddr)
		}
		return msg
	}
}
