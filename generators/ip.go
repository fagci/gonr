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
	ch      chan net.IP
	r       *rand.Rand
	max     int64
	running bool
}

func notGlobal(intip uint32) bool {
	return (intip > 0x09FFFFFF && intip < 0x0B000000) ||
		(intip > 0x643fffff && intip < 0x64800000) ||
		(intip > 0x7EFFFFFF && intip < 0x80000000) ||
		(intip > 0xA9FDFFFF && intip < 0xA9FF0000) ||
		(intip > 0xAC0FFFFF && intip < 0xAC200000) ||
		(intip > 0xBFFFFFFF && intip < 0xC0000008) ||
		(intip > 0xC00000A9 && intip < 0xC00000AC) ||
		(intip > 0xC00001FF && intip < 0xC0000300) ||
		(intip > 0xC0A7FFFF && intip < 0xC0A90000) ||
		(intip > 0xC611FFFF && intip < 0xC6140000) ||
		(intip > 0xC63363FF && intip < 0xC6336500) ||
		(intip > 0xCB0070FF && intip < 0xCB007200)
}

// Generates single WAN IP
func (g *IPGenerator) GenerateIP() net.IP {
	var intip uint32
	for {
		intip = g.r.Uint32()%0xD0000000 + 0xFFFFFF
		if notGlobal(intip) {
			continue
		}
		return Uint32ToIP(intip)
	}
}

// Generates WAN IPs to g.max count,
// passed when generator created via NewIPGenerator
func (g *IPGenerator) Generate() <-chan net.IP {
	g.running = true

	go func() {
		defer close(g.ch)
		ch := g.ch
		max := g.max

		if max >= 0 {
			var i int64
			for i = 0; i < max; i++ {
				ch <- g.GenerateIP()
			}
			return
		}

		for g.running {
			ch <- g.GenerateIP()
		}
	}()

	return g.ch
}

func (g *IPGenerator) Stop() {
	g.running = false
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
	return net.IPv4(byte(intip>>24), byte(intip>>16&0xff), byte(intip>>8&0xff), byte(intip&0xff))
}

// Creates uint32 host IPs from cidr network
func CIDRToUint32Hosts(network string) ([]uint32, error) {
	addresses, err := CIDRToUint32Addresses(network)
	if err != nil {
		return addresses, err
	}
	if len(addresses) == 1 && addresses[0]&0xff == 0 {
		return []uint32{addresses[0] + 1}, nil
	}
    arr := addresses[:0]
	for _, addr := range addresses {
		lastByte := addr & 0xff
		if lastByte != 0 && lastByte != 0xff {
			arr = append(arr, addr)
		}
	}
	return arr, nil
}

func CIDRToUint32Addresses(network string) ([]uint32, error) {
	var arr []uint32
	_, ipv4Net, err := net.ParseCIDR(network)
	if err != nil {
		return arr, err
	}
	mask := binary.BigEndian.Uint32(ipv4Net.Mask)
	start := binary.BigEndian.Uint32(ipv4Net.IP)
	finish := (start & mask) | (mask ^ 0xffffffff)
	if finish <= start {
		finish = start + 1
	}
	for i := start; i < finish; i++ {
		arr = append(arr, i)
	}
	return arr, nil
}
