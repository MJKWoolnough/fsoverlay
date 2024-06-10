package fsoverlay

import (
	"io/fs"
	"os"
)

type OS string

func (o OS) Open(name string) (fs.File, error) {
	return os.DirFS(string(o)).Open(name)
}

func (o OS) ReadFile(name string) ([]byte, error) {
	return os.DirFS(string(o)).(fs.ReadFileFS).ReadFile(name)
}

func (o OS) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.DirFS(string(o)).(fs.ReadDirFS).ReadDir(name)
}

func (o OS) Stat(name string) (fs.FileInfo, error) {
	return os.DirFS(string(o)).(fs.StatFS).Stat(name)
}
