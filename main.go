package main

import (
	"fmt"
	"github.com/ell/tostools/ipf"
	"os"
)

func main() {
	path := os.Args[1]

	ipf, err := ipf.OpenIPF(path)
	err = ipf.Parse()

	if err != nil {
		fmt.Println(err)
	}
}
