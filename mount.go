package fsoverlay

import (
	"errors"
	"io/fs"
	"slices"
	"strings"
	"sync"
)

type Mount struct {
	mu           sync.RWMutex
	mountPoints  map[string]fs.FS
	sortedPoints []string
}

func NewMount(root fs.FS) *Mount {
	return &Mount{mountPoints: map[string]fs.FS{".": root}}
}

func (m *Mount) Mount(dir string, f fs.FS) error {
	if !fs.ValidPath(dir) {
		return &fs.PathError{
			Op:   "mount",
			Path: dir,
			Err:  fs.ErrInvalid,
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if fi, err := LStat(m, dir); err != nil {
		var perr *fs.PathError
		if errors.As(err, &perr) {
			perr.Op = "mount"
		}

		return err
	} else if !fi.IsDir() {
		return &fs.PathError{
			Op:   "mount",
			Path: dir,
			Err:  fs.ErrInvalid,
		}
	}

	if _, ok := m.mountPoints[dir]; ok {
		return &fs.PathError{
			Op:   "mount",
			Path: dir,
			Err:  fs.ErrExist,
		}
	}

	m.mountPoints[dir] = f

	pos, _ := slices.BinarySearchFunc(m.sortedPoints, dir, func(a, b string) int {
		return len(a) - len(b)
	})

	m.sortedPoints = slices.Insert(m.sortedPoints, pos, dir)

	return nil
}

func (m *Mount) resolve(path string) (fs.FS, string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, mp := range m.sortedPoints {
		if strings.HasPrefix(path, mp) {
			return m.mountPoints[mp], strings.TrimPrefix(path, mp)
		}
	}

	return m.mountPoints["."], path
}

func (m *Mount) Open(path string) (fs.File, error) {
	return nil, nil
}
