/* simple profiler // o.d
 *
 * Usage:
 * import "o_o"
 *
 * ; import "github.com/charlanxcc/o_o"	// o.i 1000  <=> // o.o import 1000
 * ;;; O__o := o_o.B(100); defer O__o.E(); O__o.M(1) <=> // o.b
 *
 * ;;; o.M(1)                                        <=> // o.o
 * ;;; o.N(2, "random-name")                         <=> // o.o random-name
 * ;;; o.E()                                         <=> // o.e
 *
 * To reset:
 * o_o.Reset()
 *
 * To get summary:
 * o_o.Summary()
 */

package o_o

import (
	"bytes"
	"fmt"
	"path"
	"reflect"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
)

type point struct {
	name  string /* title of the line */
	fun   string /* file base name */
	line  int
	count int64
	et    int64
}

type O struct {
	func_ix int       /* function begin index */
	st      time.Time /* function start time */
	ot      time.Time /* last time */
}

var size int = 4096
var Enabled bool = false
var points []point = make([]point, size, size)
var dummy_o *O = &O{0, time.Now(), time.Now()}

func B(six int) *O {
	if !Enabled {
		return dummy_o
	}

	if points[six].fun == "" {
		pc := make([]uintptr, 1)
		runtime.Callers(2, pc)

		n, f, l := FuncInfo(pc[0])
		points[six].name = path.Base(n)
		points[six].fun = fmt.Sprintf("%s@%s", path.Base(n), path.Base(f))
		points[six].line = l
	}

	now := time.Now()
	return &O{
		func_ix: six,
		st:      now,
		ot:      now,
	}
}

func (o *O) N(ix int, name string) {
	if !Enabled {
		return
	}
	i := o.func_ix + ix

	if points[i].fun == "" {
		points[i].name = name

		skip := 3
		if name != "" {
			skip = 2
		}

		pc := make([]uintptr, 1)
		runtime.Callers(skip, pc)
		n, f, l := FuncInfo(pc[0])
		points[i].fun = fmt.Sprintf("%s@%s", path.Base(n), path.Base(f))
		points[i].line = l
		if ix == 0 && name == "" {
			points[i].name = points[i].fun
		}
	}

	t := time.Now()
	points[i].count = atomic.AddInt64(&points[i].count, 1)
	if ix == 0 {
		points[i].et = atomic.AddInt64(&points[i].et, int64(t.Sub(o.st)))
	} else {
		points[i].et = atomic.AddInt64(&points[i].et, int64(t.Sub(o.ot)))
		o.ot = t
	}
}

func (o *O) M(ix int) {
	if !Enabled {
		return
	}
	o.N(ix, "")
}

func (o *O) E() {
	if !Enabled {
		return
	}
	o.N(0, "")
}

func Reset() {
	if !Enabled {
		return
	}
	for i, _ := range points {
		points[i].name = ""
		points[i].fun = ""
		points[i].line = 0
		points[i].count = 0
		points[i].et = 0
	}
}

// msCutoff is cut off time in millisecond
func Summary(msCutoff int64) string {
	if !Enabled {
		return ""
	}

	var out bytes.Buffer
	out.WriteString("*** Benchmark summary: (index: name count time/op total-time func@file:line)\n")

	for i, x := range points {
		if x.count == 0 {
			continue
		}
		if msCutoff > 0 && x.et/1000000 < msCutoff {
			continue
		}

		var indent, name string = "", "........"
		if x.name != "" && strings.Contains(x.fun, x.name+"@") {
			// function
			name = x.name
		} else {
			indent = "  "
			if x.name != "" {
				name = x.name
			}
			name = (name + "        ")[:8]
		}
		out.WriteString(fmt.Sprintf(
			"%4d: %s%s: %d %.1f us/ %.3f s  %s:%d\n",
			i, indent, name, x.count,
			float64(x.et)/float64(x.count)/float64(1000),
			float64(x.et)/float64(1000000000),
			x.fun, x.line))
	}
	return out.String()
}

func (o *O) Summary(msCutoff int64) string {
	return Summary(msCutoff)
}

func FuncInfo(pc uintptr) (name, file string, line int) {
	fp := runtime.FuncForPC(pc)
	fn, l := fp.FileLine(pc)
	return fp.Name(), fn, l
}

// from function pointer
func FuncInfo2(ptr interface{}) (name, file string, line int) {
	return FuncInfo(reflect.ValueOf(ptr).Pointer())
}

func CallStack(skip, depth int) string {
	skip++
	if skip < 1 {
		skip = 1
	}
	if depth < 2 {
		depth = 2
	} else if depth > 100 {
		depth = 100
	}
	pc := make([]uintptr, depth)
	n := runtime.Callers(skip, pc)
	if n <= 0 {
		return ""
	}
	pc = pc[:n]
	frames := runtime.CallersFrames(pc)

	var out bytes.Buffer
	first := true
	for {
		if f, m := frames.Next(); !m {
			break
		} else {
			if !first {
				out.WriteString(" <= ")
			}
			first = false
			out.WriteString(fmt.Sprintf("%s:%d", f.Function, f.Line))
		}
	}

	return out.String()
}

func LogStack(prefix string) {
	fmt.Printf("%s: %s\n", prefix, CallStack(1, 10))
}

/* EOF */
