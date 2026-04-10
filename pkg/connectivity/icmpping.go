package connectivity

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// ErrICMPSocket means neither unprivileged (udp4) nor raw (ip4:icmp) ICMP could be opened.
var ErrICMPSocket = errors.New("icmp socket unavailable")

const (
	icmpIPv4Proto   = 1 // protocol number for ParseMessage (ICMPv4)
	icmpEchoTimeout = 2 * time.Second
)

var icmpEchoID uint32

// pingIPv4 sends one ICMP echo request to dest (IPv4) and waits for a matching reply.
// It tries an unprivileged datagram socket first ("udp4"), then raw "ip4:icmp".
func pingIPv4(ctx context.Context, dest net.IP) (rtt time.Duration, mode string, err error) {
	dest = dest.To4()
	if dest == nil {
		return 0, "", fmt.Errorf("icmp: need IPv4 address")
	}

	c, raw, modeLabel, err := listenIPv4ICMP()
	if err != nil {
		return 0, "", err
	}
	defer c.Close()

	id := int(atomic.AddUint32(&icmpEchoID, 1) & 0xffff)
	seq := 1
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   id,
			Seq:  seq,
			Data: []byte("netdata"),
		},
	}
	wb, err := msg.Marshal(nil)
	if err != nil {
		return 0, "", fmt.Errorf("icmp marshal: %w", err)
	}

	var wnet net.Addr = &net.UDPAddr{IP: dest}
	if raw {
		wnet = &net.IPAddr{IP: dest}
	}

	start := time.Now()
	if _, err := c.WriteTo(wb, wnet); err != nil {
		return 0, "", fmt.Errorf("icmp write: %w", err)
	}

	rb := make([]byte, 1500)
	for {
		if err := ctx.Err(); err != nil {
			return 0, "", err
		}
		if time.Since(start) > icmpEchoTimeout {
			return 0, "", fmt.Errorf("icmp echo: no reply within %s", icmpEchoTimeout)
		}

		until := time.Now().Add(250 * time.Millisecond)
		if t, ok := ctx.Deadline(); ok && t.Before(until) {
			until = t
		}
		_ = c.SetReadDeadline(until)

		n, peer, err := c.ReadFrom(rb)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			return 0, "", fmt.Errorf("icmp read: %w", err)
		}

		rm, err := icmp.ParseMessage(icmpIPv4Proto, rb[:n])
		if err != nil {
			continue
		}
		if rm.Type != ipv4.ICMPTypeEchoReply {
			continue
		}
		echo, ok := rm.Body.(*icmp.Echo)
		if !ok || echo.ID != id || echo.Seq != seq {
			continue
		}
		if got := addrIP(peer); got == nil || !got.Equal(dest) {
			continue
		}
		return time.Since(start), modeLabel, nil
	}
}

func listenIPv4ICMP() (c *icmp.PacketConn, raw bool, modeLabel string, err error) {
	pc, err := icmp.ListenPacket("udp4", "0.0.0.0")
	if err == nil {
		return pc, false, "unprivileged icmp (udp4)", nil
	}
	firstErr := err
	pc2, err2 := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err2 == nil {
		return pc2, true, "raw icmp (ip4:icmp)", nil
	}
	return nil, false, "", fmt.Errorf("%w: udp4: %v; ip4:icmp: %v", ErrICMPSocket, firstErr, err2)
}

func addrIP(a net.Addr) net.IP {
	switch v := a.(type) {
	case *net.UDPAddr:
		return v.IP
	case *net.IPAddr:
		return v.IP
	default:
		return nil
	}
}

func icmpEcho(ctx context.Context, host string) ICMPProbe {
	ip := net.ParseIP(host)
	if ip == nil {
		return ICMPProbe{Attempted: true, Target: host, OK: false, Detail: "invalid IP"}
	}
	ip = ip.To4()
	if ip == nil {
		return ICMPProbe{Attempted: true, Target: host, OK: false, Detail: "need IPv4 for icmp echo"}
	}

	rtt, mode, err := pingIPv4(ctx, ip)
	if err != nil {
		p := ICMPProbe{Attempted: true, Target: host, OK: false, Detail: err.Error()}
		if errors.Is(err, ErrICMPSocket) {
			p.NoSocket = true
		}
		return p
	}
	return ICMPProbe{
		Attempted: true,
		Target:    host,
		OK:        true,
		Detail:    fmt.Sprintf("%s, rtt %s", mode, rtt.Round(time.Millisecond)),
	}
}
