package dns

import (
	"reflect"
	"testing"
)

func TestParseResolvConf(t *testing.T) {
	in := `# comment
nameserver 127.0.0.53
nameserver 8.8.8.8
search lan example.com
domain localdomain
options edns0 trust-ad
`
	got, err := ParseResolvConf(in)
	if err != nil {
		t.Fatal(err)
	}
	want := Info{
		Nameservers: []string{"127.0.0.53", "8.8.8.8"},
		Search:      []string{"lan", "example.com"},
		Domain:      "localdomain",
		Options:     []string{"edns0", "trust-ad"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v want %+v", got, want)
	}
}

func TestParseResolvConfInvalidNameserver(t *testing.T) {
	_, err := ParseResolvConf("nameserver")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestStringSliceEqualSet(t *testing.T) {
	if !stringSliceEqualSet([]string{"a", "b"}, []string{"b", "a"}) {
		t.Fatal("order should not matter")
	}
	if stringSliceEqualSet([]string{"a"}, []string{"a", "a"}) {
		t.Fatal("multiset mismatch")
	}
	if !stringSliceEqualSet([]string{"127.0.0.53"}, []string{"127.0.0.53"}) {
		t.Fatal("equal")
	}
}
