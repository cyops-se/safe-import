package system

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type Command struct {
	name     string   `json:name`
	filename string   `json:filename`
	args     []string `json:arguments`
}

const CMD_GO = "go"
const CMD_CLAMSCAN = "clamscan"

var validCommands map[string]*Command

func Init() {
	validCommands = make(map[string]*Command, 10)
	validCommands[CMD_GO] = &Command{name: CMD_GO, filename: "/go/bin/go", args: []string{"version"}}
	validCommands[CMD_CLAMSCAN] = &Command{name: CMD_CLAMSCAN, filename: "/bin/clamdscan", args: []string{"-i", "--no-summary", "--move=/safe-import/quarantine"}}
}

func Run(command string, varargs ...string) (error, int, []byte) {
	if validCommands == nil {
		Init()
	}

	if cmd, ok := validCommands[command]; ok {

		// Add user arguments to end of static command arguments
		args := cmd.args
		for _, n := range varargs {
			args = append(args, n)
		}

		// Replace all / with \ in all arguments if running in Windows environment
		if runtime.GOOS == "windows" {
			cmd.filename = strings.ReplaceAll(cmd.filename, "/", "\\")
			for i, v := range args {
				args[i] = strings.ReplaceAll(v, "/", "\\")
			}
		}

		fmt.Println("CLAMSCAN:", cmd.filename, args)
		out, err := exec.Command(cmd.filename, args...).Output()
		if err != nil {
			return err, err.(*exec.ExitError).ExitCode(), out
		}
		return err, 0, out
	}

	err := fmt.Errorf("Invalid command: %s\n", command)
	return err, 126, []byte(command)
}

func Log(msg string) error {
	return nil
}
