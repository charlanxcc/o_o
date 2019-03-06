/* simple profiler
 * Usage:
 * import "o_o"
 *
 * o := o_o.Begin(100); defer o.End()
 *
 * o.Mark(1)
 * o.Mark0(2, "random name")
 * o.End()
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
	"runtime"
	"sync/atomic"
	"time"
)

type point struct {
	name	string		/* title of the line */
	fun		string		/* file base name */
	line	int
	count	int64
	et		int64
}

type O struct {
	func_ix int			/* function begin index */
	st		time.Time	/* function start time */
	ot		time.Time	/* last time */
}

var size int = 4096
var Enabled bool = false
var points []point = make([]point, size, size)
var dummy_o *O = &O{ 0, time.Now(), time.Now() }

func Begin(six int) *O {
	if !Enabled {
		return dummy_o
	}

	if points[six].fun == "" {
		pc := make([]uintptr, 1)
		runtime.Callers(2, pc)
		f := runtime.FuncForPC(pc[0])
		_, l := f.FileLine(pc[0])

		points[six].name = path.Base(f.Name())
		points[six].fun = points[six].name
		points[six].line = l
	}

	now := time.Now();
	return &O{
		func_ix: six,
		st: now,
		ot: now,
	}
}

func (o *O) Mark0(ix int, name string) {
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
		f := runtime.FuncForPC(pc[0])
		_, l := f.FileLine(pc[0])

		points[i].fun = path.Base(f.Name())
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

func (o *O) Mark(ix int) {
	if !Enabled {
		return
	}
	o.Mark0(ix, "")
}

func (o *O) End() {
	if !Enabled {
		return
	}
	o.Mark0(0, "")
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

func Summary() string {
	if !Enabled {
		return ""
	}

	var out bytes.Buffer
	out.WriteString("*** Benchmark summary\n")

	for i, x := range points {
		if x.count == 0 {
			continue
		}

		var indent, name string = "", "........"
		if x.name != x.fun {
			indent = "  "
		}
		if x.name != "" {
			name = x.name
		}
		out.WriteString(fmt.Sprintf(
			"  %d: %s%s: %d %.1f us/ %.3f s  %s:%d\n",
			i, indent, name, x.count,
			float64(x.et) / float64(x.count) / float64(1000),
			float64(x.et) / float64(1000000000),
			x.fun, x.line));
	}
	return out.String()
}

func (o *O) Summary() string {
	return Summary()
}

func CallStack(skip, depth int) string {
	skip++
	if (skip < 1) {
		skip = 1
	}
	if (depth < 2) {
		depth = 2
	} else if (depth > 100) {
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
