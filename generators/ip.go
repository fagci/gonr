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

var ranges = [][]uint32{
	{0xb000000, 0x643fffff},
	{0x80000000, 0xa9fdffff},
	{0x64800000, 0x7effffff},
	{0xcb007200, 0xdfffffff},
	{0xac200000, 0xbfffffff},
	{0xc0a90000, 0xc611ffff},
	{0xc6336500, 0xcb0070ff},
	{0xa9ff0000, 0xac0fffff},
	{0xc0000300, 0xc05862ff},
	{0xc0586400, 0xc0a7ffff},
	{0xc6140000, 0xc63363ff},
	{0xc0000100, 0xc00001ff},
}

var (
	sizes []uint32
	total uint32
	probs []float32
)

func init() {
	for _, r := range ranges {
		sizes = append(sizes, r[1]-r[0])
	}
	for _, s := range sizes {
		total += s
	}
	for _, s := range sizes {
		probs = append(probs, float32(s)/float32(total))
	}
}

func (g *IPGenerator) RangeRandByIndex(i int) uint32 {
	return (g.r.Uint32() % (sizes[i] + 1)) + ranges[i][0]
}

func (g *IPGenerator) GenerateIntIP() uint32 {
	var i int
	var rp float32
	p := g.r.Float32()

	for i, rp = range probs {
		if p < rp {
			break
		}
		p -= rp
	}
	return g.RangeRandByIndex(i)
}

// Generates single WAN IP
func (g *IPGenerator) GenerateIP() net.IP {
	intip := g.GenerateIntIP()
	return Uint32ToIP(intip)
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

func RandomHostsFromCIDR(network string) (hosts []net.IP, err error) {
	_, ipv4Net, err := net.ParseCIDR(network)
	if err != nil {
		return hosts, err
	}
	return RandomHostsFromNet(ipv4Net)
}

func RandomHostsFromNet(ipv4Net *net.IPNet) (hosts []net.IP, err error) {
	intHosts, err := NetToUint32Hosts(ipv4Net)
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

func NetToUint32Hosts(ipv4Net *net.IPNet) ([]uint32, error) {
	addresses, err := NetToUint32Addresses(ipv4Net)
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

// Creates uint32 host IPs from cidr network
func CIDRToUint32Hosts(network string) (hosts []uint32, err error) {
	_, ipv4Net, err := net.ParseCIDR(network)
	if err != nil {
		return hosts, err
	}
	return NetToUint32Hosts(ipv4Net)
}

func NetToUint32Addresses(ipv4Net *net.IPNet) ([]uint32, error) {
	var arr []uint32
	mask := binary.BigEndian.Uint32(ipv4Net.Mask)
	start := binary.BigEndian.Uint32(ipv4Net.IP.To4())
	finish := (start & mask) | (mask ^ 0xffffffff)
	if finish <= start {
		finish = start + 1
	}
	for i := start; i < finish; i++ {
		arr = append(arr, i)
	}
	return arr, nil
}

func CIDRToUint32Addresses(network string) ([]uint32, error) {
	var arr []uint32
	_, ipv4Net, err := net.ParseCIDR(network)
	if err != nil {
		return arr, err
	}
	return NetToUint32Addresses(ipv4Net)
}
