package fsoverlay

import (
	"os"
	"testing"
)

func TestDirFS2OS(t *testing.T) {
	const path = "some/path"

	if os, ok := DirFS2OS(os.DirFS(path)).(OS); !ok {
		t.Errorf("failed to convert DirFS to OS type")
	} else if os != path {
		t.Errorf("expecting to have OS(%s), got %s", path, os)
	}
}
