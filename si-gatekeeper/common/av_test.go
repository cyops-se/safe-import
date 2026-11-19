package common

import "testing"

func TestParseClamOutput(t *testing.T) {
	out := "/safe-import/staging/abc-eicar.com: Win.Test.EICAR_HDB-1 FOUND\r\n" +
		"/safe-import/staging/abc-eicar.com: moved to '/safe-import/quarantine/abc-eicar.com'\r\n" +
		"/safe-import/staging/odd:name.zip: Eicar-Signature FOUND\n" +
		"/safe-import/staging/odd:name.zip: moved to '/safe-import/quarantine/odd:name.zip'\n"

	infos := ParseClamOutput(out)
	if len(infos) != 2 {
		t.Fatalf("expected 2 infections, got %d: %#v", len(infos), infos)
	}

	if infos[0].Filename != "/safe-import/staging/abc-eicar.com" || infos[0].VirusName != "Win.Test.EICAR_HDB-1" {
		t.Errorf("unexpected first infection: %#v", infos[0])
	}
	if infos[0].QuarantinePath != "/safe-import/quarantine/abc-eicar.com" {
		t.Errorf("unexpected quarantine path: %q", infos[0].QuarantinePath)
	}

	// File names containing ':' must not break the parsing
	if infos[1].Filename != "/safe-import/staging/odd:name.zip" || infos[1].VirusName != "Eicar-Signature" {
		t.Errorf("unexpected second infection: %#v", infos[1])
	}
}

func TestParseClamOutputDoesNotPanic(t *testing.T) {
	inputs := []string{
		"",
		"\n\n",
		"garbage without separator",
		"FOUND",
		": FOUND",
		"/file: moved to ",
		"/file: Virus FOUND",                   // no "moved to" line
		"/a: moved to '/q/a'\n/a: Virus FOUND", // reversed order
	}
	for _, in := range inputs {
		ParseClamOutput(in) // must not panic
	}

	infos := ParseClamOutput("/file: Virus FOUND")
	if len(infos) != 1 || infos[0].QuarantinePath != "" {
		t.Errorf("unexpected result for infection without move: %#v", infos)
	}
}
