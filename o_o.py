#!/usr/bin/env python

import io
import re
import sys

# should check r_begin_code before r_begin_anno
r_ignore = re.compile(r'^.*?//\s*o.d.*$')
r_import_anno = re.compile(r'^.*?(\s*//\s*o\.i\s*(\d+).*)$')
r_import_code = re.compile(r'^.*?(\s*;+\s*import\s+"github\.com/charlanxcc/o_o"\s*//\s*o.i\s*([0-9]+).*)$')
r_begin_anno = re.compile(r'^.*?(\s*//\s*o\.b.*)$')
r_begin_code = re.compile(r'^.*?(\s*;+\s*O__o\s*:=\s*o_o\.B\(.*)$')
r_end_anno = re.compile(r'^.*?(\s*//\s*o\.e.*)$')
r_end_code = re.compile(r'^.*?(\s*;+\s*O__o\.E\(\).*)$')
r_mark_anno = re.compile(r'^.*?(\s*//\s*o\.o.*)$')
r_mark_code = re.compile(r'^.*?(\s*;+\s*O__o\.M\((\d+)\).*)$')
r_name_anno = re.compile(r'^.*?(\s*//\s*o\.o\s+(\w+).*)$')
r_name_code = re.compile(r'^.*?(\s*;+\s*O__o\.N\(\d+\s*,\s*"(\w+)"\).*)$')

# toggle between annotation <=> code
#
# ; import "gibhub.com/charlanxcc/o_o" // o.i 1000  <=> // o.i 1000
# ;;; O__o := o_o.B(100); defer O__o.E(); O__o.M(1) <=> // o.b
# ;;; O__o.M(1)                                     <=> // o.o
# ;;; O__o.N(2, "random name")                      <=> // o.o random name
# ;;; O__o.E()                                      <=> // o.e
def o_o(fin, fout, dir=0):
    out = io.BytesIO()
    found_header = False
    # 1: annotation => code
    # 2: code => annotation
    ix_begin = 0
    ix_func = 0
    ix_curr = 0
    ignore = False

    def l_e(l):
        if len(l) == 0:
            return b''
        elif len(l) == 1:
            return l
        elif l[-2:] == '\r\n':
            return l[-2:]
        elif l[-1:] == '\r' or l[-1:] == '\n':
            return l[-1:]
        else:
            return b''

    def out_import_anno(line, start, ix):
        return (line[:start].rstrip() + " // o.i " + ix + l_e(line)).encode()

    def out_import_code(line, start, ix):
        return (line[:start].rstrip() + \
                ' ; import "github.com/charlanxcc/o_o" // o.i ' + ix + \
                l_e(line)).encode()

    f = open(fin, 'r') if fin != '-' else sys.stdin
    while True:
        l = f.readline()
        if len(l) == 0:
            break

        if ignore:
            out.write(l)
            continue

        r = re.match(r_ignore, l)
        if r is not None:
            ignore = True
            out.write(l)
            continue

        r = re.match(r_import_code, l)
        if r is not None:
            found_header = True
            if dir == 0:
                dir = 2
                ix_begin = int(r.group(2))
                out.write(out_import_anno(l, r.start(1), r.group(2)))
            elif dir == 1:
                ix_begin = int(r.group(2))
                out.write(out_import_code(l, r.start(1), r.group(2)))
                pass
            else:
                ix_begin = int(r.group(2))
                out.write(out_import_anno(l, r.start(1), r.group(2)))
            ix_func = ix_begin
            ix_curr = 0
            continue
        r = re.match(r_import_anno, l)
        if r is not None:
            found_header = True
            if dir == 0:
                dir = 1
                ix_begin = int(r.group(2))
                out.write(out_import_code(l, r.start(1), r.group(2)))
            elif dir == 1:
                ix_begin = int(r.group(2))
                out.write(out_import_code(l, r.start(1), r.group(2)))
                pass
            else:
                ix_begin = int(r.group(2))
                out.write(out_import_anno(l, r.start(1), r.group(2)))
            ix_func = ix_begin
            ix_curr = 0
            continue

        if not found_header:
            out.write(l)
            continue

        r = re.match(r_begin_code, l)
        if r is None:
            r = re.match(r_begin_anno, l)
        if r is not None:
            ix_func = (ix_func + ix_curr + 9) // 10 * 10
            ix_curr = 2
            if dir == 1:
                l = l[:r.start(1)].rstrip() + " ;;; O__o := o_o.B(" + \
                        str(ix_func) + "); defer O__o.E(); O__o.M(1)" + l_e(l)
            else:
                l = l[:r.start(1)].rstrip() + " // o.b" + l_e(l)
            out.write(l.encode())
            continue

        r = re.match(r_end_code, l)
        if r is None:
            r = re.match(r_end_anno, l)
        if r is not None:
            if dir == 1:
                l = l[:r.start(1)].rstrip() + " ;;; O__o.E()" + l_e(l)
            else:
                l = l[:r.start(1)].rstrip() + " // o.e" + l_e(l)
            out.write(l.encode())
            continue

        r = re.match(r_name_code, l)
        if r is None:
            r = re.match(r_name_anno, l)
        if r is not None:
            if dir == 1:
                l = l[:r.start(1)].rstrip() + \
                        (" ;;; O__o.N(%d, \"%s\")" % (ix_curr, r.group(2))) + \
                        l_e(l)
            else:
                l = l[:r.start(1)].rstrip() + " // o.o " + r.group(2) + l_e(l)
            ix_curr += 1
            out.write(l.encode())
            continue

        r = re.match(r_mark_code, l)
        if r is None:
            r = re.match(r_mark_anno, l)
        if r is not None:
            if dir == 1:
                l = l[:r.start(1)].rstrip() + (" ;;; O__o.M(%d)" % ix_curr) + l_e(l)
            else:
                l = l[:r.start(1)].rstrip() + " // o.o" + l_e(l)
            ix_curr += 1
            out.write(l.encode())
            continue

        out.write(l)

    if f != sys.stdin:
        f.close()

    f = open(fout, 'w') if fout != '-' else sys.stdout
    f.write(out.getvalue())
    if f != sys.stdout:
        f.close()

if __name__ == "__main__":
    fin, fout, dir = None, None, 0
    for i in sys.argv[1:]:
        if i == "-a":
            dir = 2
        elif i == "-c":
            dir = 1
        elif i == "-h" or (len(i) > 1 and i[1:] == "-"):
            print("""Usage: o_o.py [-a|-c|-h] [<src.go> [<dst.go>]]

Examples:
  - update in-place, auto detecting what direction to go based on import line
  o_o.py abc.go
  - code to annotation, writing to def.go
  o_o.py -a abc.go def.go
  - in-place annotation to code
  o_o.py -c abc.go""")
            exit(0)
        elif fin is None:
            fin = i
        elif fout is None:
            fout = i
    if fin is None:
        fin = '-'
    if fout is None:
        fout = fin
    print(dir, fin, fout)
    o_o(fin, fout, dir)

# EOF
