package connectivity

import (
	"strings"
	"testing"
)

func TestSummary(t *testing.T) {
	gwOK := GatewayProbe{Attempted: true, OK: true, Detail: "ok", Address: "10.0.0.1", Iface: "eth0"}
	cases := []struct {
		rep  Report
		want string
	}{
		{Report{Gateway: GatewayProbe{Attempted: false, Detail: noIPv4DefaultRouteDetail}}, "No IPv4 default"},
		{
			Report{
				Gateway:  GatewayProbe{Attempted: true, OK: false, Address: "192.168.1.1", Iface: "eth0"},
				TCPNoDNS: Probe{OK: false},
			},
			"default gateway",
		},
		{Report{Gateway: gwOK, TCPNoDNS: Probe{OK: false}}, "routing"},
		{Report{Gateway: gwOK, TCPNoDNS: Probe{OK: true}, DNS: Probe{OK: false}}, "DNS"},
		{Report{Gateway: gwOK, TCPNoDNS: Probe{OK: true}, DNS: Probe{OK: true}, HTTPS: Probe{OK: false}}, "HTTPS"},
		{Report{Gateway: gwOK, TCPNoDNS: Probe{OK: true}, DNS: Probe{OK: true}, HTTPS: Probe{OK: true}}, "all succeeded"},
		{
			Report{
				Gateway:    gwOK,
				PublicICMP: ICMPProbe{Attempted: true, OK: false, Detail: "timeout"},
				TCPNoDNS:   Probe{OK: true},
				DNS:        Probe{OK: true},
				HTTPS:      Probe{OK: true},
			},
			"ICMP is often filtered",
		},
	}
	for _, tc := range cases {
		got := Summary(tc.rep)
		if !strings.Contains(got, tc.want) {
			t.Fatalf("Summary(%+v) = %q, want substring %q", tc.rep, got, tc.want)
		}
	}
}
