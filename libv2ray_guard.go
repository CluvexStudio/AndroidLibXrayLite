package libv2ray

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

var (
	crashMu   sync.Mutex
	crashFile *os.File
)

func SetCrashFile(path string) string {
	crashMu.Lock()
	defer crashMu.Unlock()
	if st, err := os.Stat(path); err == nil && st.Size() > 0 {
		_ = os.Rename(path, path+".prev")
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err.Error()
	}
	if err := debug.SetCrashOutput(f, debug.CrashOptions{}); err != nil {
		_ = f.Close()
		return err.Error()
	}
	if crashFile != nil {
		_ = crashFile.Close()
	}
	crashFile = f
	return ""
}

func recoverCall(where string, err *error, onPanic func()) {
	r := recover()
	if r == nil {
		return
	}
	stack := debug.Stack()
	log.Printf("libv2ray: %s panicked: %v\n%s", where, r, stack)
	crashMu.Lock()
	if crashFile != nil {
		fmt.Fprintf(crashFile, "%s recovered panic in %s: %v\n%s\n", time.Now().Format(time.RFC3339), where, r, stack)
	}
	crashMu.Unlock()
	if err != nil {
		*err = fmt.Errorf("%s: core panic: %v", where, r)
	}
	if onPanic != nil {
		onPanic()
	}
}
