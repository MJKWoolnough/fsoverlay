// A package to combine multiple fs.FS implementations into a single tree.
package fsoverlay

import (
	"errors"
	"io/fs"
	"path"
	"time"
)

type Overlay []fs.FS

func (o Overlay) Open(path string) (fs.File, error) {
	for _, ofs := range o {
		f, err := ofs.Open(path)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}

		return f, err
	}

	return nil, &fs.PathError{
		Op:   "open",
		Path: path,
		Err:  fs.ErrNotExist,
	}
}

func (o Overlay) ReadFile(name string) ([]byte, error) {
	for _, ofs := range o {
		data, err := fs.ReadFile(ofs, name)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}

		return data, err
	}

	return nil, &fs.PathError{
		Op:   "readfile",
		Path: name,
		Err:  fs.ErrNotExist,
	}
}

func (o Overlay) Stat(name string) (fs.FileInfo, error) {
	for _, ofs := range o {
		fi, err := fs.Stat(ofs, name)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}

		return fi, err
	}

	return nil, &fs.PathError{
		Op:   "stat",
		Path: name,
		Err:  fs.ErrNotExist,
	}
}

func (o Overlay) Sub(dir string) (fs.FS, error) {
	var p Overlay

	for _, ofs := range o {
		if pfs, err := fs.Sub(ofs, dir); err == nil {
			p = append(p, pfs)
		}
	}

	return p, nil
}

type readLink interface {
	Readlink(string) (string, error)
}

func (o Overlay) Readlink(name string) (string, error) {
	for _, ofs := range o {
		target, err := Readlink(ofs, name)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}

		return target, err
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
		fi, err := LStat(ofs, name)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}

		return fi, err
	}

	return nil, &fs.PathError{
		Op:   "lstat",
		Path: name,
		Err:  fs.ErrNotExist,
	}
}

type deAsFI struct {
	fs.DirEntry
}

func (deAsFI) Size() int64 {
	return -1
}

func (d deAsFI) Mode() fs.FileMode {
	return d.Type()
}

func (deAsFI) ModTime() time.Time {
	return time.Time{}
}

func (d deAsFI) Sys() any {
	return d.DirEntry
}

func LStat(f fs.FS, name string) (fs.FileInfo, error) {
	if lf, ok := f.(lstat); ok {
		return lf.LStat(name)
	}

	dir, base := path.Split(name)

	dirEntries, err := fs.ReadDir(f, dir)
	if err != nil {
		return nil, &fs.PathError{Op: "lstat", Path: name, Err: errors.Unwrap(err)}
	}

	for _, de := range dirEntries {
		if de.Name() == base {
			if de.Type()&fs.ModeSymlink != 0 {
				if fi, ok := de.(fs.FileInfo); ok {
					return fi, nil
				}

				return deAsFI{DirEntry: de}, nil
			} else {
				fi, err := fs.Stat(f, name)
				if err != nil {
					err = &fs.PathError{Op: "lstat", Path: name, Err: errors.Unwrap(err)}
				}

				return fi, err
			}
		}
	}

	return nil, &fs.PathError{Op: "lstat", Path: name, Err: fs.ErrNotExist}
}

func Readlink(f fs.FS, name string) (string, error) {
	if rl, ok := f.(readLink); ok {
		return rl.Readlink(name)
	}

	return "", &fs.PathError{
		Op:   "readlink",
		Path: name,
		Err:  fs.ErrInvalid,
	}
}
