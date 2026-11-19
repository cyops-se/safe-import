package system

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

// Command is a whitelisted external command
type Command struct {
	Name     string
	Filename string
	Args     []string
}

const CMD_GO = "go"
const CMD_CLAMSCAN = "clamscan"

// Exit codes returned by Run when the command could not be executed at all
const (
	ExitCodeNotFound      = 2   // Treated as a scan failure (same as clamdscan's "error" exit code)
	ExitCodeInvalid       = 126 // The command is not whitelisted
	ExitCodeUnknownFailed = 127 // The command failed for another reason
)

var (
	validCommands map[string]*Command
	initOnce      sync.Once
)

func getenv(name, def string) string {
	if v := strings.TrimSpace(os.Getenv(name)); v != "" {
		return v
	}
	return def
}

// QuarantineDir returns the directory where infected files are moved
func QuarantineDir() string {
	return getenv("SI_QUARANTINE_DIR", "/safe-import/quarantine")
}

// Init sets up the whitelisted commands. Paths can be overridden with SI_GO_BIN and SI_CLAMDSCAN.
// Set SI_CLAMDSCAN_STREAM=1 when clamd runs on another host/container and can not read the
// files directly; clamdscan then streams the file content to clamd.
func Init() {
	initOnce.Do(func() {
		clamArgs := []string{"-i", "--no-summary", "--move=" + QuarantineDir()}
		if getenv("SI_CLAMDSCAN_STREAM", "0") == "1" {
			clamArgs = append(clamArgs, "--stream")
		}

		validCommands = map[string]*Command{
			CMD_GO:       {Name: CMD_GO, Filename: getenv("SI_GO_BIN", "/go/bin/go"), Args: []string{"version"}},
			CMD_CLAMSCAN: {Name: CMD_CLAMSCAN, Filename: getenv("SI_CLAMDSCAN", "/bin/clamdscan"), Args: clamArgs},
		}
	})
}

// Run executes a whitelisted command with the static arguments followed by varargs.
// It returns the error (if any), the exit code and the standard output of the command.
func Run(command string, varargs ...string) (error, int, []byte) {
	Init()

	cmd, ok := validCommands[command]
	if !ok {
		return fmt.Errorf("invalid command: %s", command), ExitCodeInvalid, []byte(command)
	}

	// Copy the static arguments so that the shared command definition is never modified
	args := make([]string, 0, len(cmd.Args)+len(varargs))
	args = append(args, cmd.Args...)
	args = append(args, varargs...)
	filename := cmd.Filename

	// Replace all / with \ in all arguments if running in Windows environment
	if runtime.GOOS == "windows" {
		filename = strings.ReplaceAll(filename, "/", "\\")
		for i, v := range args {
			args[i] = strings.ReplaceAll(v, "/", "\\")
		}
	}

	out, err := exec.Command(filename, args...).Output()
	if err != nil {
		var exitErr *exec.ExitError
		var pathErr *fs.PathError
		switch {
		case errors.As(err, &exitErr):
			return err, exitErr.ExitCode(), out
		case errors.As(err, &pathErr), errors.Is(err, exec.ErrNotFound):
			return err, ExitCodeNotFound, out
		default:
			return err, ExitCodeUnknownFailed, out
		}
	}

	return nil, 0, out
}
