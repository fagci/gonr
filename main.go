package main

import (
	"flag"
	"fmt"

	"github.com/fagci/gonr/generators"
)

var max int64

func init() {
	flag.Int64Var(&max, "n", 5, "number of random IP to generate; negative values means infinite generation")
}

func main() {
	flag.Parse()

	generator := generators.NewIPGenerator(0, max)
	for ip := range generator.Generate() {
		fmt.Println(ip)
	}
}
