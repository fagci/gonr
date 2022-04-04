package generators

import (
	"fmt"
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
