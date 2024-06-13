package fsoverlay

import (
	"io/fs"
	"path"
	"strings"
)

const (
	readPerm     = 0o444
	maxRedirects = 255
)

type resolver struct {
	fs                 fs.FS
	fullPath, path     string
	cutAt              int
	redirectsRemaining uint8
}

func (o Overlay) resolve(path string) (string, error) {
	if !fs.ValidPath(path) {
		return "", fs.ErrInvalid
	}

	r := resolver{
		fs:                 o,
		fullPath:           path,
		redirectsRemaining: maxRedirects,
	}

	if err := r.resolve(true); err != nil {
		return "", err
	}

	return r.fullPath, nil
}

func (r *resolver) resolve(resolveLast bool) error {
	root, err := LStat(r.fs, ".")
	if err != nil {
		return err
	}

	curr := root

	for r.path != "" {
		if curr.Mode()&readPerm == 0 {
			return fs.ErrPermission
		} else if !curr.IsDir() {
			return fs.ErrInvalid
		} else if name := r.splitOffNextPart(); isEmptyName(name) {
			continue
		} else if curr, err = LStat(r.fs, name); err != nil {
			return err
		} else if r.isDone(resolveLast) {
			break
		} else if curr.Mode()&fs.ModeSymlink == 0 {
			continue
		} else if err = r.handleSymlink(name); err != nil {
			return err
		}

		curr = root
	}

	return nil
}

func (r *resolver) splitOffNextPart() string {
	slashPos := strings.Index(r.path, "/")

	if slashPos == -1 {
		r.path = ""
		r.cutAt = len(r.fullPath)
	} else {
		r.path = r.path[slashPos+1:]
		r.cutAt += slashPos + 1
	}

	return r.fullPath[:r.cutAt]
}

func (r *resolver) handleSymlink(file string) error {
	r.redirectsRemaining--
	if r.redirectsRemaining == 0 {
		return fs.ErrInvalid
	}

	target, err := Readlink(r.fs, path.Join(r.path, file))
	if err != nil {
		return err
	}

	if strings.HasPrefix(target, "/") {
		r.fullPath = path.Clean(target)[1:]
	} else if r.path == "" {
		r.fullPath = path.Join(r.fullPath[:r.cutAt], target)
	} else {
		r.fullPath = path.Join(r.fullPath[:r.cutAt-len(file)-1], target, r.path)
	}

	r.path = r.fullPath
	r.cutAt = 0

	return nil
}

func (r *resolver) isDone(resolveLast bool) bool {
	return r.path == "" && !resolveLast
}

func isEmptyName(name string) bool {
	return name == "" || name == "."
}
