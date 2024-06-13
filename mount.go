package fsoverlay

import (
	"io/fs"
	"slices"
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
