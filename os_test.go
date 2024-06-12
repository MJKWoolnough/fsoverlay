package fsoverlay

import (
	"io"
	"io/fs"
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

func TestOSOpen(t *testing.T) {
	tmp, _, testFile, testSymlink := createTestOS(t)

	for n, filename := range [...]string{testFile, testSymlink} {
		if f, err := tmp.Open(filename); err != nil {
			t.Errorf("test %d: unexpected error opening file: %s", n+1, err)
		} else if contents, err := io.ReadAll(f); err != nil {
			t.Errorf("test %d: error reading file contents: %s", n+1, err)
		} else if readContents := string(contents); readContents != testContents {
			t.Errorf("test %d: expected to read %q, read %q", n+1, testContents, readContents)
		}
	}
}

func TestOSReadFile(t *testing.T) {
	tmp, _, testFile, testSymlink := createTestOS(t)

	for n, filename := range [...]string{testFile, testSymlink} {
		if contents, err := tmp.ReadFile(filename); err != nil {
			t.Errorf("test %d: error reading file contents: %s", n+1, err)
		} else if readContents := string(contents); readContents != testContents {
			t.Errorf("test %d: expected to read %q, read %q", n+1, testContents, readContents)
		}
	}
}

func TestOSReadDir(t *testing.T) {
	tmp, testDir, testFile, testSymlink := createTestOS(t)

	if entries, err := tmp.ReadDir(testDir); err != nil {
		t.Errorf("unexpected error during ReadDir: %s", err)
	} else if len(entries) != 2 {
		t.Errorf("expecting to read 2 entries, got %d", len(entries))
	} else if e1, e2, t1, t2 := entries[0].Name(), entries[1].Name(), filepath.Base(testFile), filepath.Base(testSymlink); e1 != t1 || e2 != t2 {
		t.Errorf("expecting to read []string{%q, %q}, read []string{%q, %q}", t1, t2, e1, e2)
	}
}

func TestOSStat(t *testing.T) {
	tmp, _, testFile, testSymlink := createTestOS(t)

	for n, filename := range [...]string{testFile, testSymlink} {
		if fi, err := tmp.Stat(filename); err != nil {
			t.Errorf("test %d: error reading file contents: %s", n+1, err)
		} else if fn, tn := fi.Name(), filepath.Base(filename); fn != tn {
			t.Errorf("test %d: expected to stat file %s, got %s", n+1, tn, fn)
		} else if m := fi.Mode(); m&fs.ModePerm != m {
			t.Errorf("test %d: expected to stat file", n+1)
		}
	}
}

func TestOSLStat(t *testing.T) {
	tmp, _, testFile, testSymlink := createTestOS(t)

	if fi, err := tmp.LStat(testFile); err != nil {
		t.Errorf("test %d: error reading file contents: %s", 1, err)
	} else if fn, tn := fi.Name(), filepath.Base(testFile); fn != tn {
		t.Errorf("test %d: expected to stat file %s, got %s", 1, tn, fn)
	} else if m := fi.Mode(); m&fs.ModePerm != m {
		t.Errorf("test %d: expected to stat file", 1)
	}

	if fi, err := tmp.LStat(testSymlink); err != nil {
		t.Errorf("test %d: error reading file contents: %s", 2, err)
	} else if fn, tn := fi.Name(), filepath.Base(testSymlink); fn != tn {
		t.Errorf("test %d: expected to stat file %s, got %s", 2, tn, fn)
	} else if m := fi.Mode(); m&fs.ModeSymlink == 0 {
		t.Errorf("test %d: expected to stat symlink", 2)
	}
}

func TestOSReadLink(t *testing.T) {
	tmp, _, testFile, testSymlink := createTestOS(t)

	if target, err := tmp.ReadLink(testSymlink); err != nil {
		t.Errorf("unexpected error readlink symlink target: %s", err)
	} else if expectedTarget := filepath.Join(string(tmp), testFile); target != expectedTarget {
		t.Errorf("expecting to read link %q, got %q", expectedTarget, target)
	}
}
