package common

import (
	"log"
	"strings"

	"github.com/cyops-se/safe-import/si-gatekeeper/system"
	"github.com/cyops-se/safe-import/si-gatekeeper/types"
)

// Scan scans a file with ClamAV. A non-nil error means the file must NOT be used: either
// an infection was found (exit code 1) or the scan could not be completed (any other code).
func Scan(localfile string) (error, int, []types.InfectionInfo) {
	err, exitcode, out := system.Run(system.CMD_CLAMSCAN, localfile)
	if err == nil {
		return nil, exitcode, nil
	}

	var infos []types.InfectionInfo
	if exitcode == 1 {
		log.Printf("clamdscan returns: %s", string(out))
		infos = ParseClamOutput(string(out))
	}

	return err, exitcode, infos
}

// ParseClamOutput parses the output of 'clamdscan -i --no-summary --move=<dir>', which
// for each infected file prints:
//
//	/path/to/file: Virus.Name FOUND
//	/path/to/file: moved to '/quarantine/file'
//
// Unknown lines are ignored. File names may contain ':' so the last ": " is used as separator.
func ParseClamOutput(text string) []types.InfectionInfo {
	var infos []types.InfectionInfo
	index := make(map[string]int) // file name -> index in infos

	text = strings.ReplaceAll(text, "\r", "")
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasSuffix(line, " FOUND") {
			sep := strings.LastIndex(line, ": ")
			if sep <= 0 {
				continue
			}
			filename := line[:sep]
			virus := strings.TrimSpace(strings.TrimSuffix(line[sep+2:], " FOUND"))
			index[filename] = len(infos)
			infos = append(infos, types.InfectionInfo{Filename: filename, OriginalPath: filename, VirusName: virus})
			continue
		}

		if sep := strings.Index(line, ": moved to "); sep > 0 {
			filename := line[:sep]
			dest := strings.Trim(strings.TrimSpace(line[sep+len(": moved to "):]), "'\"")
			if i, ok := index[filename]; ok {
				infos[i].QuarantinePath = dest
			}
		}
	}

	return infos
}
