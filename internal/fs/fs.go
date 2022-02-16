package fs

import (
	"embed"
	"io"
	"os"
)

type FS struct {
	Core           embed.FS
	InjectOverride *embed.FS
}

func (f *FS) Open(path string) (io.ReadCloser, error) {
	if f.InjectOverride != nil {
		file, err := f.InjectOverride.Open(path)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		if err == nil {
			return file, nil
		}
	}

	return f.Core.Open(path)
}

func (f *FS) Stat(path string) (os.FileInfo, error) {
	if f.InjectOverride != nil {
		file, err := f.InjectOverride.Open(path)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		if err == nil {
			return file.Stat()
		}
	}

	file, err := f.Core.Open(path)
	if err != nil {
		return nil, err
	}
	return file.Stat()
}
