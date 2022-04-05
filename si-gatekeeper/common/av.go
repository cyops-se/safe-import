package common

import (
	"log"
	"strings"

	"github.com/cyops-se/safe-import/si-gatekeeper/system"
	"github.com/cyops-se/safe-import/si-gatekeeper/types"
)

func Scan(localfile string) (error, int, []types.InfectionInfo) {

	var infos []types.InfectionInfo
	err, exitcode, out := system.Run(system.CMD_CLAMSCAN, localfile)
	if err != nil {
		// fmt.Printf("CLAM exited with status code: %d, out: %s", exitcode, string(out))
		text := string(out)
		if exitcode == 1 {
			// Infections are reported from ClamAV as two lines per found infection separated by \n (and possibly \r)
			// Lets report each infection individually
			info := strings.ReplaceAll(text, "\r", "") // First normalize the text by removing all \r

			log.Printf("clamavd returns: %s", info)

			lines := strings.Split(info, "\n")
			for i := 0; i < len(lines)-1; i += 2 {
				parts := strings.Split(lines[i], ":")
				entry := &types.InfectionInfo{}
				entry.VirusName = parts[len(parts)-1]
				entry.Filename = parts[len(parts)-2]

				parts = strings.Split(lines[i+1], " moved to ")
				entry.OriginalPath = parts[0]
				entry.QuarantinePath = parts[1]

				infos = append(infos, *entry)
			}
		}
	}

	return err, exitcode, infos
}
