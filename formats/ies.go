package formats

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"errors"
	_ "fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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
	Unknown [5]byte
	Order   uint8
}

type nodes []node

func (n nodes) Len() int      { return len(n) }
func (n nodes) Swap(i, j int) { n[i], n[j] = n[j], n[i] }

func (n nodes) Less(i, j int) bool {
	return n[i].Order < n[j].Order
}

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

	fileName := filepath.Base(ies.File.Name()) + ".csv"

	filePath := filepath.Join(path)
	os.MkdirAll(filePath, 0777)

	csvfile, err := os.Create(filepath.Join(filePath, fileName))

	if err != nil {
		return err
	}

	defer csvfile.Close()

	writer := csv.NewWriter(csvfile)

	header := make([]string, len(ies.Nodes))
	for i, head := range ies.Nodes {
		header[i] = head.NameOne
	}

	writer.Write(header)

	for _, row := range ies.Rows {
		r := make([]string, len(ies.Nodes))
		for j, node := range ies.Nodes {
			r[j] = row[node.NameOne]
		}

		writer.Write(r)
	}

	writer.Flush()
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
		Unknown [5]byte
		Order   uint8
	}
	var strNodes []node
	var intNodes []node

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
			Order:   tmp.Order,
		}

		if n.FmtType == 0 {
			intNodes = append(intNodes, n)
		} else {
			strNodes = append(strNodes, n)
		}

	}

	var nnodes []node

	sort.Sort(nodes(intNodes))
	sort.Sort(nodes(strNodes))
	nnodes = append(intNodes, strNodes...)

	ies.Nodes = nnodes

	return nil
}

func (ies *IES) parseRows() error {
	// sort.Sort(nodes(ies.Nodes))
	// if ies.Nodes[1].NameOne == "ClassID" {
	// 	ies.Nodes[0], ies.Nodes[1] = ies.Nodes[1], ies.Nodes[0]
	// }

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
