package formats

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

type IES struct {
	File *os.File
	Key  int

	Header headerSection
	Data   dataSection
}

type headerSection struct {
	fileName   string
	dataOffset uint32
	resOffset  uint32
	eofOffset  uint32
	resPtr     uint32
	dataPtr    uint32
	numRows    uint16
	numFormats uint16
}

type dataSection struct {
	numRows   uint16
	unknown_3 uint16
	unknown_0 uint16
	order     uint16
}

func OpenIES(filepath string) (*IES, error) {
	var ies IES

	file, err := os.Open(filepath)
	if err != nil {
		return &ies, err
	}

	ies.File = file
	ies.Key = 0x01

	return &ies, nil
}

func (ies *IES) Parse() error {
	err := ies.parseHeader()
	return err
}

func (ies *IES) Decompress(path string) error {
	return nil
}

func (ies *IES) parseHeader() error {
	var head headerSection

	nameBuf := make([]byte, 128)
	dataOffsetBuf := make([]byte, 4)
	resOffsetBuf := make([]byte, 4)
	eofOffsetBuf := make([]byte, 4)
	numRowsBuf := make([]byte, 2)
	numFormatsBuf := make([]byte, 2)

	_, err := ies.File.ReadAt(nameBuf, 0)
	if err != nil {
		return err
	}

	_, err = ies.File.ReadAt(dataOffsetBuf, 0x84)
	if err != nil {
		return err
	}

	_, err = ies.File.ReadAt(resOffsetBuf, 0x88)
	if err != nil {
		return err
	}

	_, err = ies.File.ReadAt(eofOffsetBuf, 0x8C)
	if err != nil {
		return err
	}

	_, err = ies.File.ReadAt(numRowsBuf, 0x92)
	if err != nil {
		return err
	}

	head.fileName = strings.TrimRight(string(nameBuf), "\x00")
	head.dataOffset = readInt32(dataOffsetBuf)
	head.resOffset = readInt32(resOffsetBuf)
	head.eofOffset = readInt32(eofOffsetBuf)

	head.resPtr = head.eofOffset - head.resOffset
	head.dataPtr = head.resPtr - head.dataOffset

	if hasRowsBuf == 0x01 {
		head.numRows = 0
		head.numFormats = 0
	} else {
		_, err = ies.File.ReadAt(numFormatsBuf, 0x94)
		if err != nil {
			return err
		}

		head.numFormats = numFormatsBuf
	}

	fmt.Printf("%+v", head)
	return nil
}

func (ies *IES) parseDataSection() error {
	//ies.File.Seek(ies.Header.dataPtr, 0)
	return nil

}

func (ies *IES) parseTableSection() error {
	return nil
}

func readInt32(data []byte) (r uint32) {
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.LittleEndian, &r)
	return
}
