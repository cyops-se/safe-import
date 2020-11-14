package common

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/cyops-se/safe-import/si-gatekeeper/system"
	"github.com/cyops-se/safe-import/si-gatekeeper/types"
)

func Scan(localfile string) (error, int, []types.InfectionInfo) {

	err, exitcode, out := system.Run(system.CMD_CLAMSCAN, localfile)
	if err != nil {
		fmt.Printf("CLAM exited with status code: %d, out: %s", exitcode, string(out))
		text := string(out)
		if exitcode == 1 {
			// Infections are reported from ClamAV as two lines per found infection separated by \n (and possibly \r)
			// Lets report each infection individually
			info := strings.ReplaceAll(text, "\r", "") // First normalize the text by removing all \r
			parts := strings.Split(info, "\n")
			for i := 0; i < len(parts)-1; i += 2 {
				if runtime.GOOS == "windows" {
					// svc.LogInfection(strings.Split(parts[i], ":")[2], parts[i+1])
					fmt.Println(strings.Split(parts[i], ":")[2], parts[i+1])
				} else {
					// svc.LogInfection(strings.Split(parts[i], ":")[1], parts[i+1])
					fmt.Println(strings.Split(parts[i], ":")[2], parts[i+1])
				}
			}
		}
	}

	return err, exitcode, nil
}
