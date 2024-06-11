package fsoverlay

import (
	"os"
	"path/filepath"
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

const testContents = "my data"

func createTestOS(t *testing.T) (OS, string, string, string) {
	t.Helper()

	tmp := t.TempDir()
	testDir := "testDir"
	testFile := filepath.Join(testDir, "testFile")
	testSymlink := filepath.Join(testDir, "testSymlink")

	err := os.Mkdir(filepath.Join(tmp, testDir), 0o755)
	if err != nil {
		t.Fatalf("failed to create test dir: %s", err)
	}

	f, err := os.Create(filepath.Join(tmp, testFile))
	if err != nil {
		t.Fatalf("failed to create test file: %s", err)
	}

	if _, err := f.WriteString(testContents); err != nil {
		t.Fatalf("failed to write to test file: %s", err)
	}

	if err := f.Close(); err != nil {
		t.Fatalf("failed to close test file: %s", err)
	}

	if err := os.Symlink(filepath.Join(tmp, testFile), filepath.Join(tmp, testSymlink)); err != nil {
		t.Fatalf("failed to create test symlink: %s", err)
	}

	return OS(tmp), testDir, testFile, testSymlink
}
