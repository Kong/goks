package fs

import (
	"embed"
	"io"
	"os"
)

type FS struct {
	EmbedFS embed.FS
}

func (f *FS) Open(path string) (io.ReadCloser, error) {
	return f.EmbedFS.Open(path)
}

func (f *FS) Stat(path string) (os.FileInfo, error) {
	file, err := f.EmbedFS.Open(path)
	if err != nil {
		return nil, err
	}
	return file.Stat()
}
