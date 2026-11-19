package common

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRelativeToRoot(t *testing.T) {
	root := filepath.FromSlash("/safe-import/outer")
	ok := map[string]string{
		"/safe-import/outer/example.com/file.zip":    filepath.FromSlash("example.com/file.zip"),
		"/safe-import/outer/a/../b/file":             filepath.FromSlash("b/file"),
		"/safe-import/outer//double//slash/file.txt": filepath.FromSlash("double/slash/file.txt"),
	}
	for in, want := range ok {
		got, err := RelativeToRoot(root, filepath.FromSlash(in))
		if err != nil || got != want {
			t.Errorf("RelativeToRoot(%q) = %q, %v; want %q", in, got, err, want)
		}
	}

	bad := []string{
		"/safe-import/outer",
		"/safe-import/outer/../inner/file",
		"/safe-import/outer/../../etc/shadow",
		"/etc/shadow",
		"/safe-import/outerx/file",
	}
	for _, in := range bad {
		if got, err := RelativeToRoot(root, filepath.FromSlash(in)); err == nil {
			t.Errorf("RelativeToRoot(%q) = %q, expected error", in, got)
		}
	}
}

func TestCopyToStagingAndMoveToInner(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SI_STAGING_DIR", filepath.Join(dir, "staging"))

	src := filepath.Join(dir, "outer", "file.bin")
	os.MkdirAll(filepath.Dir(src), 0755)
	content := []byte("some downloaded content")
	if err := os.WriteFile(src, content, 0644); err != nil {
		t.Fatal(err)
	}

	staged, sum, err := CopyToStaging(src)
	if err != nil {
		t.Fatal(err)
	}
	expected := sha256.Sum256(content)
	if sum != hex.EncodeToString(expected[:]) {
		t.Errorf("unexpected sha256 %s", sum)
	}

	// Changing the source after staging must not affect the staged copy
	os.WriteFile(src, []byte("malicious replacement"), 0644)
	if data, _ := os.ReadFile(staged); string(data) != string(content) {
		t.Errorf("staged copy changed: %q", data)
	}

	dest := filepath.Join(dir, "inner", "sub", "file.bin")
	if err := MoveToInner(staged, dest); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(dest); string(data) != string(content) {
		t.Errorf("inner file has wrong content: %q", data)
	}
	if info, err := os.Lstat(dest); err != nil || info.Mode()&os.ModeSymlink != 0 {
		t.Errorf("inner file must be a regular file, got %v %v", info, err)
	}
	if _, err := os.Stat(staged); !os.IsNotExist(err) {
		t.Errorf("staged file should be gone after move")
	}
}

func TestCopyToStagingRejectsSymlinks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("creating symlinks requires privileges on Windows")
	}
	dir := t.TempDir()
	t.Setenv("SI_STAGING_DIR", filepath.Join(dir, "staging"))

	secret := filepath.Join(dir, "secret")
	os.WriteFile(secret, []byte("secret"), 0600)
	link := filepath.Join(dir, "link")
	if err := os.Symlink(secret, link); err != nil {
		t.Fatal(err)
	}

	if _, _, err := CopyToStaging(link); err == nil {
		t.Error("expected symlink to be rejected")
	}
}
