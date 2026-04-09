package routes

import (
	"net"
	"testing"
)

func TestParseProcHexIPv4(t *testing.T) {
	ip, err := parseProcHexIPv4("00000000")
	if err != nil {
		t.Fatal(err)
	}
	if !ip.Equal(net.IPv4zero) {
		t.Fatalf("got %v want 0.0.0.0", ip)
	}
	// Little-endian 0x010011AC -> 172.17.0.1
	ip, err = parseProcHexIPv4("010011AC")
	if err != nil {
		t.Fatal(err)
	}
	want := net.IPv4(172, 17, 0, 1)
	if !ip.Equal(want) {
		t.Fatalf("got %v want %v", ip, want)
	}
}

func TestParseProcNetRoute(t *testing.T) {
	sample := `Iface	Destination	Gateway 	Flags	RefCnt	Use	Metric	Mask		MTU	Window	IRTT
eth0	00000000	010011AC	0003	0	0	100	00000000	0	0	0
docker0	000011AC	00000000	0001	0	0	0	0000FFFF	0	0	0
`
	rs, err := parseProcNetRoute(sample)
	if err != nil {
		t.Fatal(err)
	}
	if len(rs) != 2 {
		t.Fatalf("got %d routes want 2", len(rs))
	}
	if rs[0].Iface != "eth0" || rs[0].Destination != "0.0.0.0" {
		t.Fatalf("first route: %+v", rs[0])
	}
	if rs[0].Gateway != "172.17.0.1" {
		t.Fatalf("gateway got %q", rs[0].Gateway)
	}
}

func TestFormatDestination(t *testing.T) {
	if got := FormatDestination("0.0.0.0", "0.0.0.0"); got != "default" {
		t.Fatalf("got %q want default", got)
	}
	if got := FormatDestination("172.17.0.0", "255.255.0.0"); got != "172.17.0.0" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatRTFFlags(t *testing.T) {
	if got := FormatRTFFlags("0003"); got != "UG" {
		t.Fatalf("0003 got %q want UG", got)
	}
	if got := FormatRTFFlags("0001"); got != "U" {
		t.Fatalf("0001 got %q want U", got)
	}
}
