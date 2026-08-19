package agent

import (
	"net"
	"reflect"
	"testing"
)

func TestFamiliesFromIPsDetectsNATIPv4AndGlobalIPv6(t *testing.T) {
	addresses := []net.IP{net.ParseIP("192.168.0.138"), net.ParseIP("2001:db8::23"), net.ParseIP("fe80::1"), net.ParseIP("127.0.0.1")}
	if got := familiesFromIPs(addresses); !reflect.DeepEqual(got, []int{4, 6}) {
		t.Fatalf("families = %v, want [4 6]", got)
	}
}
