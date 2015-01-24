package main

import (
	"fmt"
	"github.com/ell/tostools/formats"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Invalid args. tostools <type> <input> <out dir>")
		return
	}

	t := os.Args[1]
	in := os.Args[2]
	out, err := filepath.Abs(os.Args[3])

	if err != nil {
		fmt.Println(err)
		return
	}

	f := ParseArgs(t, in)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = f.Parse()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = f.Decompress(out)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func ParseArgs(ftype, in string) formats.TOSFormat {
	t := strings.ToUpper(ftype)

	switch t {
	case "IES":
		ies, err := formats.OpenIES(in)
		if err != nil {
			panic(err)
		}
		return ies
	case "IPF":
		ipf, err := formats.OpenIPF(in)
		if err != nil {
			panic(err)
		}
		return ipf
	}

	fmt.Println("Unsupported File Format. Choices are: ies, ipf")
	return nil
}
