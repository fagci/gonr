package generators

import (
	"bytes"
	"fmt"
	"sort"
	"testing"
)

func TestIPGen(t *testing.T) {
	const N = 5
	var i int

	ipGen := NewIPGenerator(1, 0)
	for range ipGen.Generate() {
		i++
	}
	if i != 0 {
		t.Error("IP count is not zero")
	}

	ipGen = NewIPGenerator(1, N)
	for ip := range ipGen.Generate() {
		fmt.Println(ip.String())
		i++
	}
	if i != N {
		t.Error("Wrong IP count")
	}
}
func TestRandomHostsFromCIDR(t *testing.T) {
	network := "192.168.0.1/24"
	randomHosts, err := RandomHostsFromCIDR(network)
	if err != nil {
		t.Error(err)
	}
	sort.Slice(randomHosts, func(i, j int) bool {
		return bytes.Compare(randomHosts[i], randomHosts[j]) < 0
	})
	if len(randomHosts) != 254 {
		t.Error("Wrong IPs count")
	}
	if randomHosts[0].String() != "192.168.0.1" {
		t.Error("First addr not in range")
	}
	if randomHosts[len(randomHosts)-1].String() != "192.168.0.254" {
		t.Error("Last addr not in range")
	}
}
