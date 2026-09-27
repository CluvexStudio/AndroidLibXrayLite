package libv2ray

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func panicky() (delay int64, err error) {
	defer recoverCall("panicky", &err, func() { delay = -1 })
	var m map[string]int
	m["boom"] = 1
	return 5, nil
}

func TestRecoverCallTurnsAPanicIntoAnError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "core-crash.txt")
	if msg := SetCrashFile(path); msg != "" {
		t.Fatal(msg)
	}
	delay, err := panicky()
	if err == nil || delay != -1 {
		t.Fatalf("got delay=%d err=%v, want -1 and an error", delay, err)
	}
	if !strings.Contains(err.Error(), "panicky: core panic") {
		t.Errorf("error does not say where: %v", err)
	}
	body, _ := os.ReadFile(path)
	if !strings.Contains(string(body), "recovered panic in panicky") {
		t.Errorf("crash file does not record the recovered panic:\n%s", body)
	}
}

func TestSetCrashFileKeepsThePreviousRun(t *testing.T) {
	path := filepath.Join(t.TempDir(), "core-crash.txt")
	if err := os.WriteFile(path, []byte("panic: from the last run\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if msg := SetCrashFile(path); msg != "" {
		t.Fatal(msg)
	}
	prev, err := os.ReadFile(path + ".prev")
	if err != nil || !strings.Contains(string(prev), "from the last run") {
		t.Fatalf("previous crash not kept: %q, %v", prev, err)
	}
	if st, _ := os.Stat(path); st.Size() != 0 {
		t.Errorf("new crash file should start empty, has %d bytes", st.Size())
	}
}
