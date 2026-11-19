package system

import (
	"sync"
	"testing"
)

func TestRunRejectsUnknownCommands(t *testing.T) {
	err, code, _ := Run("rm", "-rf", "/")
	if err == nil || code != ExitCodeInvalid {
		t.Errorf("expected invalid command error, got %v, %d", err, code)
	}
}

func TestRunMissingBinaryFailsClosed(t *testing.T) {
	t.Setenv("SI_CLAMDSCAN", "/does/not/exist/clamdscan")
	validCommands = nil
	initOnce = sync.Once{}
	defer func() { validCommands = nil; initOnce = sync.Once{} }()

	err, code, _ := Run(CMD_CLAMSCAN, "/some/file")
	if err == nil {
		t.Fatal("expected an error when the scanner is missing")
	}
	if code == 0 || code == 1 {
		t.Errorf("a missing scanner must not be reported as clean (0) or infected (1), got %d", code)
	}
}

func TestRunDoesNotModifyStaticArguments(t *testing.T) {
	Init()
	before := len(validCommands[CMD_CLAMSCAN].Args)
	Run(CMD_CLAMSCAN, "/a")
	Run(CMD_CLAMSCAN, "/b")
	if after := len(validCommands[CMD_CLAMSCAN].Args); after != before {
		t.Errorf("static arguments modified: %d -> %d", before, after)
	}
}
