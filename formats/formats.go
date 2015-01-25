package formats

import (
	"bytes"
	"encoding/binary"
)

type TOSFormat interface {
	Parse() error
	Decompress(string) error
}

func readInt32(data []byte) (r uint32) {
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.LittleEndian, &r)
	return
}

func readInt16(data []byte) (r uint16) {
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.LittleEndian, &r)
	return
}

func readXorString(data []byte, key byte) (r string) {

	strip_garbo := func(r rune) bool {
		if r == 0 || r == 3 || r == 1 || r == 2 {
			return true
		} else {
			return false
		}
	}

	d := bytes.TrimFunc(data, strip_garbo)
	rBuf := make([]byte, len(d))

	for i := 0; i < len(d); i++ {
		rBuf[i] = d[i] ^ key
	}

	return string(rBuf)
}
