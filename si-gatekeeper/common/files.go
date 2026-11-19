package common

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Directories used by the gatekeeper. They can be overridden with environment variables.
func OuterDir() string   { return getenv("SI_OUTER_DIR", "/safe-import/outer") }
func InnerDir() string   { return getenv("SI_INNER_DIR", "/safe-import/inner") }
func StagingDir() string { return getenv("SI_STAGING_DIR", "/safe-import/staging") }

func getenv(name, def string) string {
	if v := strings.TrimSpace(os.Getenv(name)); v != "" {
		return v
	}
	return def
}

// RelativeToRoot returns the path of file relative to root, or an error if file is not
// located inside root (e.g. "../" tricks or absolute paths elsewhere). si-outer is the
// least trusted component, so file names it reports must never be trusted blindly.
func RelativeToRoot(root string, file string) (string, error) {
	root = filepath.Clean(root)
	file = filepath.Clean(file)

	rel, err := filepath.Rel(root, file)
	if err != nil {
		return "", fmt.Errorf("path %s is not inside %s: %w", file, root, err)
	}
	if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("path %s is not inside %s", file, root)
	}

	return rel, nil
}

// CopyToStaging copies src to a new, uniquely named file in the staging directory and
// returns its path and SHA-256. Symlinks are not followed for the source file, so that
// si-outer can not trick the gatekeeper into copying arbitrary files.
func CopyToStaging(src string) (string, string, error) {
	info, err := os.Lstat(src)
	if err != nil {
		return "", "", err
	}
	if !info.Mode().IsRegular() {
		return "", "", fmt.Errorf("%s is not a regular file", src)
	}

	if err := os.MkdirAll(StagingDir(), 0750); err != nil {
		return "", "", err
	}

	in, err := os.Open(src)
	if err != nil {
		return "", "", err
	}
	defer in.Close()

	suffix := make([]byte, 8)
	if _, err := rand.Read(suffix); err != nil {
		return "", "", err
	}
	staged := filepath.Join(StagingDir(), hex.EncodeToString(suffix)+"-"+filepath.Base(src))

	out, err := os.OpenFile(staged, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return "", "", err
	}

	hash := sha256.New()
	if _, err := io.Copy(io.MultiWriter(out, hash), in); err != nil {
		out.Close()
		os.Remove(staged)
		return "", "", err
	}
	if err := out.Close(); err != nil {
		os.Remove(staged)
		return "", "", err
	}

	return staged, hex.EncodeToString(hash.Sum(nil)), nil
}

// MoveToInner moves a scanned file from the staging area to dest (inside the inner directory).
// Falls back to copy + remove if the directories are on different file systems.
func MoveToInner(staged string, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	tmp := dest + ".si-tmp"
	os.Remove(tmp)

	if err := os.Rename(staged, tmp); err != nil {
		if err := copyFile(staged, tmp); err != nil {
			os.Remove(tmp)
			return err
		}
		os.Remove(staged)
	}

	os.Remove(dest) // Remove a previous version (or symlink created by older versions)
	return os.Rename(tmp, dest)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
