package generators

import (
	crypto_rand "crypto/rand"
	"encoding/binary"
	"errors"
	"math/rand"
	"net"
	"strings"
)

type IPGenerator struct {
	ch  chan net.IP
	r   *rand.Rand
	max int64
}

// Generates single WAN IP
func (g *IPGenerator) GenerateIP() net.IP {
	var intip uint32
	for {
		intip = g.r.Uint32()%0xD0000000 + 0xFFFFFF
		if (intip >= 0x0A000000 && intip <= 0x0AFFFFFF) ||
			(intip >= 0x64400000 && intip <= 0x647FFFFF) ||
			(intip >= 0x7F000000 && intip <= 0x7FFFFFFF) ||
			(intip >= 0xA9FE0000 && intip <= 0xA9FEFFFF) ||
			(intip >= 0xAC100000 && intip <= 0xAC1FFFFF) ||
			(intip >= 0xC0000000 && intip <= 0xC0000007) ||
			(intip >= 0xC00000AA && intip <= 0xC00000AB) ||
			(intip >= 0xC0000200 && intip <= 0xC00002FF) ||
			(intip >= 0xC0A80000 && intip <= 0xC0A8FFFF) ||
			(intip >= 0xC6120000 && intip <= 0xC613FFFF) ||
			(intip >= 0xC6336400 && intip <= 0xC63364FF) ||
			(intip >= 0xCB007100 && intip <= 0xCB0071FF) {
			continue
		}
		break
	}
	return Uint32ToIP(intip)
}

// Generates WAN IPs to g.max count,
// passed when generator created via NewIPGenerator
func (g *IPGenerator) Generate() <-chan net.IP {
	go func() {
		defer close(g.ch)

		if g.max >= 0 {
			var i int64
			for i = 0; g.max < 0 || i < g.max; i++ {
				g.ch <- g.GenerateIP()
			}
			return
		}

		for {
			g.ch <- g.GenerateIP()
		}
	}()

	return g.ch
}

// Creates new WAN IP generator with capacity of channel and max count of IPs to generate via  Generate()
func NewIPGenerator(capacity int, max int64) *IPGenerator {
	return &IPGenerator{
		ch:  make(chan net.IP, capacity),
		r:   NewCryptoRandom(),
		max: max,
	}
}

func NewCryptoRandom() *rand.Rand {
	b := make([]byte, 8)
	_, err := crypto_rand.Read(b)
	if err != nil {
		panic("Cryptorandom seed failed: " + err.Error())
	}
	return rand.New(rand.NewSource(int64(binary.LittleEndian.Uint64(b))))
}

func RandomHostsFromCIDRGen(network string) <-chan net.IP {
	ch := make(chan net.IP)
	hosts, err := RandomHostsFromCIDR(network)
	if err != nil {
		panic(err)
	}
	go func() {
		defer close(ch)
		for _, host := range hosts {
			ch <- host
		}
	}()
	return ch
}

func RandomHostsFromListGen(list []string) <-chan net.IP {
	ch := make(chan net.IP)
	hosts, err := RandomHostsFromList(list)
	if err != nil {
		panic(err)
	}
	go func() {
		defer close(ch)
		for _, host := range hosts {
			ch <- host
		}
	}()
	return ch
}

func RandomHostsFromCIDR(network string) ([]net.IP, error) {
	var hosts []net.IP
	intHosts, err := CIDRToUint32Hosts(network)
	if err != nil {
		return hosts, err
	}
	r := NewCryptoRandom()
	r.Shuffle(len(intHosts), func(i, j int) {
		intHosts[i], intHosts[j] = intHosts[j], intHosts[i]
	})
	for _, intHost := range intHosts {
		hosts = append(hosts, Uint32ToIP(intHost))
	}
	return hosts, nil
}

func RandomHostsFromList(list []string) ([]net.IP, error) {
	var hosts []net.IP
	r := NewCryptoRandom()
	r.Shuffle(len(list), func(i, j int) {
		list[i], list[j] = list[j], list[i]
	})
	for _, line := range list {
		if strings.IndexRune(line, '/') > 0 {
			_hosts, err := RandomHostsFromCIDR(line)
			if err != nil {
				return hosts, err
			}
			hosts = append(hosts, _hosts...)
		} else {
			ip := net.ParseIP(line)
			if ip == nil {
				return hosts, errors.New("wrong IP: " + line)
			}
			hosts = append(hosts, ip)
		}
	}
	return hosts, nil
}

func Uint32ToIP(intip uint32) net.IP {
	return net.IPv4(byte(intip>>24), byte(intip>>16), byte(intip>>8), byte(intip))
}

// Creates uint32 host IPs from cidr network
func CIDRToUint32Hosts(network string) ([]uint32, error) {
	var arr []uint32
	_, ipv4Net, err := net.ParseCIDR(network)
	if err != nil {
		return arr, err
	}
	mask := binary.BigEndian.Uint32(ipv4Net.Mask)
	start := binary.BigEndian.Uint32(ipv4Net.IP)
	finish := (start & mask) | (mask ^ 0xffffffff)
	for i := start + 1; i < finish; i++ {
		arr = append(arr, i)
	}
	return arr, nil
}
