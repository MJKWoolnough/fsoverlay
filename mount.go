package fsoverlay

import (
	"io/fs"
)

type Mount struct {
	mountPoints  map[string]fs.FS
	sortedPoints []string
}

func NewMount(root fs.FS) *Mount {
	return &Mount{mountPoints: map[string]fs.FS{".": root}}
}
