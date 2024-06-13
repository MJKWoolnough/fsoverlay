package fsoverlay

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"reflect"
	"strings"
)

func DirFS2OS(os fs.FS) fs.FS {
	v := reflect.ValueOf(os)

	if t := v.Type(); t.PkgPath() == "os" && t.Name() == "dirFS" {
		return OS(v.String())
	}

	return os
}

type OS string

func (o OS) Open(name string) (fs.File, error) {
	return os.DirFS(string(o)).Open(name)
}

func (o OS) ReadFile(name string) ([]byte, error) {
	return fs.ReadFile(os.DirFS(string(o)), name)
}

func (o OS) ReadDir(name string) ([]fs.DirEntry, error) {
	return fs.ReadDir(os.DirFS(string(o)), name)
}

func (o OS) Stat(name string) (fs.FileInfo, error) {
	return fs.Stat(os.DirFS(string(o)), name)
}

func (o OS) LStat(name string) (fs.FileInfo, error) {
	pname, err := join(string(o), name)
	if err != nil {
		return nil, &fs.PathError{
			Op:   "lstat",
			Path: name,
			Err:  err,
		}
	}

	fi, err := os.Lstat(pname)
	if err != nil {
		var perr *os.PathError
		if errors.As(err, &perr) {
			perr.Path = name
		}

		return nil, err
	}

	return fi, nil
}

func join(base, name string) (string, error) {
	if base == "" {
		return "", ErrEmptyRoot
	}

	combined := path.Join(base, name)
	toCheck := strings.TrimPrefix(combined, "/")

	if !fs.ValidPath(toCheck) {
		return "", fs.ErrInvalid
	}

	return combined, nil
}

func (o OS) Readlink(name string) (string, error) {
	pname, err := join(string(o), name)
	if err != nil {
		return "", &fs.PathError{
			Op:   "readlink",
			Path: name,
			Err:  err,
		}
	}

	link, err := os.Readlink(pname)
	if err != nil {
		var perr *os.PathError
		if errors.As(err, &perr) {
			perr.Path = name
		}

		return "", err
	}

	return link, nil
}

var ErrEmptyRoot = errors.New("invalid root directory")
