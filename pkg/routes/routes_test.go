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

func TestPrimaryIPv4DefaultRouteFromRoutes(t *testing.T) {
	list := []Route{
		{Iface: "eth0", Destination: "0.0.0.0", Mask: "0.0.0.0", Gateway: "10.0.0.2", Metric: "200"},
		{Iface: "wlan0", Destination: "0.0.0.0", Mask: "0.0.0.0", Gateway: "192.168.1.1", Metric: "100"},
	}
	got, err := primaryIPv4DefaultRouteFromRoutes(list)
	if err != nil {
		t.Fatal(err)
	}
	if got.Gateway != "192.168.1.1" || got.Iface != "wlan0" || got.Metric != 100 {
		t.Fatalf("got %+v", got)
	}
	_, err = primaryIPv4DefaultRouteFromRoutes([]Route{
		{Destination: "10.0.0.0", Mask: "255.0.0.0", Gateway: "0.0.0.0", Metric: "0"},
	})
	if err != ErrNoIPv4DefaultRoute {
		t.Fatalf("got err %v want ErrNoIPv4DefaultRoute", err)
	}
}
