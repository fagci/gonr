package main

import (
	"flag"
	"fmt"

	"github.com/fagci/gonr/generators"
)

var max int64
var network string

func init() {
	flag.Int64Var(&max, "n", 5, "number of random IP to generate; negative values means infinite generation")
    flag.StringVar(&network, "net", "", "Network in CIDR notation to generate random hosts from")
}

func main() {
	flag.Parse()

    if network != "" {
        for ip := range generators.RandomHostsFromCIDRGen(network) {
            fmt.Println(ip)
        }
        return
    }

	generator := generators.NewIPGenerator(0, max)
	for ip := range generator.Generate() {
		fmt.Println(ip)
	}
}
