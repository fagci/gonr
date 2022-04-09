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
		t.Errorf("Wrong IPs count (%d)", len(randomHosts))
	}
    first := randomHosts[0].String()
	if first != "192.168.0.1" {
        t.Errorf("First addr not in range: %s", first)
	}
    last := randomHosts[len(randomHosts)-1].String()
	if last != "192.168.0.254" {
        t.Errorf("Last addr not in range: %s", last)
	}
}

func TestRandomHostsFromCIDRComplicated(t *testing.T) {
	network := "192.168.0.33/22"
	needCount := 254 * 4
	randomHosts, err := RandomHostsFromCIDR(network)
	if err != nil {
		t.Error(err)
	}
	sort.Slice(randomHosts, func(i, j int) bool {
		return bytes.Compare(randomHosts[i], randomHosts[j]) < 0
	})
	if len(randomHosts) != needCount {
		t.Errorf("Wrong IPs count (%d, need %d)", len(randomHosts), needCount)
	}
    first := randomHosts[0].String()
	if first != "192.168.0.1" {
        t.Errorf("First addr not in range: %s", first)
	}
    last := randomHosts[len(randomHosts)-1].String()
	if last != "192.168.3.254" {
        t.Errorf("Last addr not in range: %s", last)
	}
}

func TestSingleHostFromCIDR(t *testing.T) {
	network := "192.168.0.0/32"
	randomHosts, err := RandomHostsFromCIDR(network)
	if err != nil {
		t.Error(err)
	}
	if len(randomHosts) != 1 {
		t.Errorf("Wrong IPs count (%d)", len(randomHosts))
	}
	if randomHosts[0].String() != "192.168.0.1" {
		t.Error("First addr not in range")
	}
	network = "192.168.0.1/32"
	randomHosts, err = RandomHostsFromCIDR(network)
	if err != nil {
		t.Error(err)
	}
	if len(randomHosts) != 1 {
		t.Errorf("Wrong IPs count (%d)", len(randomHosts))
	}
	if randomHosts[0].String() != "192.168.0.1" {
		t.Error("First addr not in range")
	}
}

func TestRandomHostsFromList(t *testing.T) {
	list := []string{"127.0.0.1", "192.168.0.1/24"}
	randomHosts, err := RandomHostsFromList(list)
	if err != nil {
		t.Error(err)
	}
	sort.Slice(randomHosts, func(i, j int) bool {
		return bytes.Compare(randomHosts[i], randomHosts[j]) < 0
	})
	if len(randomHosts) != 255 {
		t.Errorf("Wrong IPs count (%d)", len(randomHosts))
	}
	if randomHosts[0].String() != "127.0.0.1" {
		t.Error("First addr not in range")
	}
	if randomHosts[len(randomHosts)-1].String() != "192.168.0.254" {
		t.Error("Last addr not in range")
	}
}
