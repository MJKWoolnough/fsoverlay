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

	if firstError == nil {
		return nil, &fs.PathError{
			Op:   "open",
			Path: path,
			Err:  ErrNoFSs,
		}
	}

	return nil, firstError
}

func (o Overlay) ReadFile(name string) ([]byte, error) {
	var firstError error

	for _, ofs := range o {
		data, err := fs.ReadFile(ofs, name)
		if errors.Is(err, fs.ErrNotExist) {
			if firstError == nil {
				firstError = err
			}

			continue
		}

		return data, err
	}

	if firstError == nil {
		return nil, &fs.PathError{
			Op:   "readfile",
			Path: name,
			Err:  ErrNoFSs,
		}
	}

	return nil, firstError
}

func (o Overlay) Stat(name string) (fs.FileInfo, error) {
	var firstError error

	for _, ofs := range o {
		fi, err := fs.Stat(ofs, name)
		if errors.Is(err, fs.ErrNotExist) {
			if firstError == nil {
				firstError = err
			}

			continue
		}

		return fi, err
	}

	if firstError == nil {
		return nil, &fs.PathError{
			Op:   "stat",
			Path: name,
			Err:  ErrNoFSs,
		}
	}

	return nil, firstError
}

type readLink interface {
	Readlink(string) (string, error)
}

func (o Overlay) Readlink(name string) (string, error) {
	for _, ofs := range o {
		if rl, ok := ofs.(readLink); ok {
			link, err := rl.Readlink(name)
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}

			return link, err
		}
	}

	return "", &fs.PathError{
		Op:   "readlink",
		Path: name,
		Err:  fs.ErrNotExist,
	}
}

type lstat interface {
	LStat(string) (fs.FileInfo, error)
}

func (o Overlay) LStat(name string) (fs.FileInfo, error) {
	for _, ofs := range o {
		if rl, ok := ofs.(lstat); ok {
			fi, err := rl.LStat(name)
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}

			return fi, err
		}
	}

	return nil, &fs.PathError{
		Op:   "lstat",
		Path: name,
		Err:  fs.ErrNotExist,
	}
}

// Errors.
var (
	ErrNoFSs = errors.New("no overlays")
)
