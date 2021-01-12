package content

import (
	file2 "github.com/enorith/supports/file"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

type uploadFile struct {
	header *multipart.FileHeader
	file   multipart.File
}

func (uf *uploadFile) Save(dist string) error {
	dir := filepath.Dir(dist)

	if dir == dist {
		dist = filepath.Join(dist, uf.Filename())
	}

	exist, e := file2.PathExists(dir)
	if e != nil {
		return e
	}
	if !exist {
		e = os.MkdirAll(dir, 0775)
		if e != nil {
			return e
		}
	}

	d, e := os.OpenFile(dist, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0775)
	if e != nil {
		return e
	}

	defer d.Close()
	file, e := uf.Open()

	if e != nil {
		return e
	}

	defer uf.Close()
	_, e = io.Copy(d, file)
	if e != nil {
		return e
	}

	return nil
}

func (uf *uploadFile) Open() (multipart.File, error) {
	if uf.file != nil {
		return uf.file, nil
	}

	return uf.header.Open()
}

func (uf *uploadFile) Close() error {
	if uf.file != nil {
		return uf.file.Close()
	}

	return nil
}
func (uf *uploadFile) Filename() string {
	return uf.header.Filename
}
