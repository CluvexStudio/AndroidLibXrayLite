package libv2ray

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
)

func RuntimeReport(limit int) (report string) {
	defer recoverCall("RuntimeReport", nil, func() { report = "" })
	if limit <= 0 || limit > 40 {
		limit = 12
	}

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	var b strings.Builder
	fmt.Fprintf(&b, "goroutines %d · threads %d · GOMAXPROCS %d\n",
		runtime.NumGoroutine(), runtime.NumCPU(), runtime.GOMAXPROCS(0))
	fmt.Fprintf(&b, "heap %s live · %s from the OS · stacks %s · %d GCs · %.1f%% of CPU in GC\n",
		humanBytes(ms.HeapAlloc), humanBytes(ms.HeapSys+ms.StackSys), humanBytes(ms.StackInuse),
		ms.NumGC, ms.GCCPUFraction*100)
	fmt.Fprintf(&b, "%s allocated in total, %s freed\n", humanBytes(ms.TotalAlloc), humanBytes(ms.TotalAlloc-ms.HeapAlloc))

	for _, g := range goroutineGroups(limit) {
		fmt.Fprintf(&b, "\n%d ×\n", g.count)
		for _, f := range g.frames {
			b.WriteString("    ")
			b.WriteString(f)
			b.WriteByte('\n')
		}
	}
	return b.String()
}

type goroutineGroup struct {
	count  int
	frames []string
}

const framesKept = 14

func goroutineGroups(limit int) []goroutineGroup {
	var records []runtime.StackRecord
	for i := 0; i < 3; i++ {
		n, _ := runtime.GoroutineProfile(nil)
		records = make([]runtime.StackRecord, n+32)
		if n, ok := runtime.GoroutineProfile(records); ok {
			records = records[:n]
			break
		}
		records = nil
	}
	if records == nil {
		return nil
	}

	groups := make(map[string]*goroutineGroup, len(records))
	for i := range records {
		frames := describe(records[i].Stack())
		key := strings.Join(frames, "\n")
		if g := groups[key]; g != nil {
			g.count++
		} else {
			groups[key] = &goroutineGroup{count: 1, frames: frames}
		}
	}

	out := make([]goroutineGroup, 0, len(groups))
	for _, g := range groups {
		out = append(out, *g)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].count > out[j].count })
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

func describe(pcs []uintptr) []string {
	if len(pcs) == 0 {
		return []string{"(no stack)"}
	}
	var all []string
	frames := runtime.CallersFrames(pcs)
	for {
		f, more := frames.Next()
		if f.Function != "" {
			name := strings.TrimPrefix(f.Function, "github.com/xtls/xray-core/")
			if f.File != "" {
				name = fmt.Sprintf("%s (%s:%d)", name, f.File[strings.LastIndexByte(f.File, '/')+1:], f.Line)
			}
			all = append(all, name)
		}
		if !more {
			break
		}
	}
	if len(all) == 0 {
		return []string{"(no stack)"}
	}

	out := []string{all[0]}
	rest := all[1:]
	for len(rest) > 0 && isRuntimeFrame(rest[0]) {
		rest = rest[1:]
	}
	if len(rest) == 0 {
		rest = all[1:]
	}
	for _, f := range rest {
		if len(out) >= framesKept {
			break
		}
		out = append(out, f)
	}
	return out
}

func isRuntimeFrame(name string) bool {
	for _, p := range [...]string{"runtime.", "internal/runtime", "internal/abi", "internal/poll", "sync."} {
		if strings.HasPrefix(name, p) {
			return true
		}
	}
	return false
}

func humanBytes(n uint64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := uint64(unit), 0
	for v := n / unit; v >= unit && exp < 3; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGT"[exp])
}
