package formats

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strconv"
	_ "strings"
)

type IES struct {
	File     *os.File
	Key      byte
	Header   headerSection
	DataInfo dataInfo
	Nodes    []node
	Rows     []map[string]string
}

type headerSection struct {
	Name          string
	OffsetHintA   uint32
	OffsetHintB   uint32
	FileSize      uint32
	OffsetColumns uint32
	OffsetRows    uint32
}

type dataInfo struct {
	ValOne    uint16
	Rows      uint16
	Columns   uint16
	ColInt    uint16
	ColString uint16
}

type node struct {
	NameOne string
	NameTwo string
	FmtType byte
}
type nodes []node

func OpenIES(filepath string) (*IES, error) {
	var ies IES

	file, err := os.Open(filepath)
	if err != nil {
		return &ies, err
	}

	ies.File = file
	ies.Key = byte(0x01)

	return &ies, nil
}

func (ies *IES) Parse() error {
	err := ies.parseHeader()
	if err != nil {
		return err
	}

	err = ies.parseDataSection()
	if err != nil {
		return err
	}

	err = ies.parseFormatsSection()
	if err != nil {
		return err
	}

	err = ies.parseRows()
	if err != nil {
		return err
	}

	return nil
}

func (ies *IES) Decompress(path string) error {
	return nil
}

func (ies *IES) parseHeader() error {

	type header struct {
		Name        [128]byte
		Val1        uint32
		OffsetHintA uint32
		OffsetHintB uint32
		FileSize    uint32
	}

	var head header

	headBuf := make([]byte, 144)
	_, err := ies.File.Read(headBuf)
	err = binary.Read(bytes.NewBuffer(headBuf), binary.LittleEndian, &head)

	if err != nil {
		return err
	}

	hs := headerSection{
		Name:          readXorString(head.Name[:], ies.Key),
		OffsetHintA:   head.OffsetHintA,
		OffsetHintB:   head.OffsetHintB,
		FileSize:      head.FileSize,
		OffsetColumns: head.FileSize - (head.OffsetHintA + head.OffsetHintB),
		OffsetRows:    head.FileSize - head.OffsetHintB,
	}

	ies.Header = hs

	return nil
}

func (ies *IES) parseDataSection() error {
	var d dataInfo

	emptyCheckBuf := make([]byte, 2)
	_, err := ies.File.ReadAt(emptyCheckBuf, 0x92)

	if err != nil {
		return err
	}

	if readInt16(emptyCheckBuf) == 0x01 {
		return errors.New("not supported yet")
	}

	ies.File.Seek(144, 0)
	dataBuf := make([]byte, 12)
	_, err = ies.File.Read(dataBuf)

	if err != nil {
		return err
	}

	err = binary.Read(bytes.NewBuffer(dataBuf), binary.LittleEndian, &d)
	if err != nil {
		return err
	}

	ies.DataInfo = d

	return nil

}

func (ies *IES) parseFormatsSection() error {
	type fileNode struct {
		NameOne [64]byte
		NameTwo [64]byte
		FmtType byte
		Unknown [6]byte
	}
	var nodes []node

	offset := int64(ies.Header.OffsetColumns)

	_, err := ies.File.Seek(offset, 0)
	if err != nil {
		return err
	}

	for i := 0; i < int(ies.DataInfo.Columns); i++ {
		var tmp fileNode

		buf := make([]byte, 136)
		_, err := ies.File.Read(buf)

		err = binary.Read(bytes.NewBuffer(buf), binary.LittleEndian, &tmp)
		if err != nil {
			return err
		}

		n := node{
			NameOne: readXorString(tmp.NameOne[:], ies.Key),
			NameTwo: readXorString(tmp.NameTwo[:], ies.Key),
			FmtType: tmp.FmtType,
		}

		nodes = append(nodes, n)
	}

	ies.Nodes = nodes

	return nil
}

func (ies *IES) parseRows() error {
	offset := int64(ies.Header.OffsetRows)
	ies.File.Seek(offset, 0)

	for curRow := 0; curRow < int(ies.DataInfo.Rows); curRow++ {
		row := make(map[string]string)

		indexBuf := make([]byte, 4)
		optionalBuf := make([]byte, 2)

		ies.File.Read(indexBuf)
		ies.File.Read(optionalBuf)

		optionalStringBuf := make([]byte, readInt16(optionalBuf))
		ies.File.Read(optionalStringBuf)

		for i := 0; i < int(ies.DataInfo.ColInt); i++ {
			intBuf := make([]byte, 4)
			ies.File.Read(intBuf)

			row[ies.Nodes[i].NameOne] = strconv.Itoa(int(readInt16(intBuf)))
		}

		for j := 0; j < int(ies.DataInfo.ColString); j++ {
			strSizeBuf := make([]byte, 2)
			ies.File.Read(strSizeBuf)

			strBuf := make([]byte, readInt16(strSizeBuf))
			ies.File.Read(strBuf)

			row[ies.Nodes[j+int(ies.DataInfo.ColInt)].NameOne] = readXorString(strBuf, ies.Key)
		}

		ies.File.Seek(int64(ies.DataInfo.ColString), 1)
		ies.Rows = append(ies.Rows, row)
	}

	return nil
}
