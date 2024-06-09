package fsoverlay

import (
	"errors"
	"io/fs"
)

type Overlay []fs.FS

func (o Overlay) Open(path string) (fs.File, error) {
	var firstError error

	for _, ofs := range o {
		f, err := ofs.Open(path)
		if errors.Is(err, fs.ErrNotExist) {
			if firstError == nil {
				firstError = err
			}

			continue
		}

		return f, err
	}

	return nil, firstError
}
