package formats

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type IPF struct {
	File  *os.File
	Meta  meta
	Files []fileMeta
}

type meta struct {
	Files  uint16
	Offset uint32
	Flag   uint16
}

type fileMeta struct {
	Nsize   uint16
	Crc     uint32
	Zsize   uint32
	Size    uint32
	Offset  uint32
	Csize   uint16
	Comment string
	Name    string
}

func OpenIPF(file string) (*IPF, error) {
	var ipf IPF

	f, err := os.Open(file)
	if err != nil {
		return &ipf, err
	}

	ipf.File = f

	return &ipf, nil
}

func (ipf *IPF) Parse() error {
	meta, err := getMeta(ipf.File)
	if err != nil {
		return err
	}

	ipf.Meta = meta

	err = ipf.GetFileList()
	if err != nil {
		return err
	}

	return nil
}

func (ipf *IPF) Decompress(basePath string) error {
	var err error

	for _, file := range ipf.Files {
		err = file.Decompress(basePath, ipf)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ipf *IPF) GetFileList() error {
	files := []fileMeta{}

	type file struct {
		Nsize  uint16
		Crc    uint32
		Zsize  uint32
		Size   uint32
		Offset uint32
		Csize  uint16
	}

	offset := int64(ipf.Meta.Offset)

	for i := 0; i < int(ipf.Meta.Files); i++ {
		var f file
		_, err := ipf.File.Seek(offset, 0)

		data := make([]byte, 20)
		_, err = ipf.File.Read(data)

		err = binary.Read(bytes.NewBuffer(data), binary.LittleEndian, &f)
		offset = offset + 20

		ipf.File.Seek(offset, 0)
		commentData := make([]byte, f.Csize)
		_, err = ipf.File.ReadAt(commentData, offset)

		offset = offset + int64(f.Csize)
		ipf.File.Seek(offset, 0)
		nameData := make([]byte, f.Nsize)
		_, err = ipf.File.ReadAt(nameData, offset)

		offset = offset + int64(f.Nsize)
		ipf.File.Seek(offset, 0)

		m := fileMeta{
			Nsize:   f.Nsize,
			Crc:     f.Crc,
			Zsize:   f.Zsize,
			Size:    f.Size,
			Offset:  f.Offset,
			Csize:   f.Csize,
			Comment: string(commentData),
			Name:    string(nameData),
		}

		if err != nil {
			return err
		}

		files = append(files, m)
	}

	ipf.Files = files

	return nil
}

func getMeta(file *os.File) (meta, error) {
	var m meta

	stat, _ := os.Stat(file.Name())
	offset := int64(24)
	start := stat.Size() - offset

	buf := make([]byte, offset)
	_, err := file.ReadAt(buf, start)

	err = binary.Read(bytes.NewBuffer(buf), binary.LittleEndian, &m)

	if err != nil {
		fmt.Println(err)
		return m, err
	}

	return m, nil
}

func (fm *fileMeta) Decompress(rootPath string, ipf *IPF) error {
	ignore := []string{".mp3", ".fsb", ".jpg", ".JPG"}

	buf := make([]byte, fm.Zsize)
	_, err := ipf.File.ReadAt(buf, int64(fm.Offset))
	if err != nil {
		return err
	}

	b := bytes.NewReader(buf)
	r := flate.NewReader(b)

	path := filepath.Dir(fm.Name)
	ext := filepath.Ext(fm.Name)
	fileName := filepath.Base(fm.Name)

	fullPath := filepath.Join(rootPath, filepath.Base(ipf.File.Name()), path)
	os.MkdirAll(fullPath, 0777)

	fullerPath := filepath.Join(fullPath, fileName)
	out, err := os.Create(fullerPath)

	if err != nil {
		return err
	}

	for _, e := range ignore {
		if ext == e {
			io.Copy(out, b)
		} else {
			io.Copy(out, r)
		}
	}
	r.Close()

	return nil
}
