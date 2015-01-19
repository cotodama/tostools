package main

import (
	"fmt"
	"github.com/ell/tostools/ipf"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Invalid args. tostools <input> <out dir>")
		return
	}

	in := os.Args[1]
	out, err := filepath.Abs(os.Args[2])

	fmt.Printf("%s %s", in, out)

	if err != nil {
		fmt.Println(err)
		return
	}

	ipf, err := ipf.OpenIPF(in)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = ipf.Parse()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = ipf.Decompress(out)
	if err != nil {
		fmt.Println(err)
		return
	}
}
