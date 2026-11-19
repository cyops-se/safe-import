package services

import (
	"path/filepath"
	"testing"
)

func TestIsInside(t *testing.T) {
	dir := filepath.FromSlash("/safe-import/inner")
	cases := map[string]bool{
		"/safe-import/inner/example.com/file.zip":   true,
		"/safe-import/inner/a/../b":                 true,
		"/safe-import/inner":                        false,
		"/safe-import/inner/../outer/file":          false,
		"/safe-import/innerx/file":                  false,
		"/etc/passwd":                               false,
		"/safe-import/inner/../../../../etc/passwd": false,
	}
	for in, want := range cases {
		if got := isInside(dir, filepath.FromSlash(in)); got != want {
			t.Errorf("isInside(%q) = %v, want %v", in, got, want)
		}
	}
}
