package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/elves/upgrade-scripts-for-0.17/fix"
	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/parse"
)

var (
	rewrite = flag.Bool("w", false, "rewrite files")
	lambda  = flag.Bool("lambda", false, "migrate lambda syntax")
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		process("[stdin]", os.Stdin, os.Stdout)
	} else {
		for _, arg := range args {
			f, err := os.OpenFile(arg, os.O_RDWR, 0)
			if err != nil {
				diag.ShowError(os.Stderr, err)
				continue
			}
			w := os.Stdout
			if *rewrite {
				w = f
			}
			process(arg, f, w)
		}
	}
}

func process(name string, r io.Reader, w io.Writer) {
	code, err := io.ReadAll(r)
	if err != nil {
		diag.ShowError(os.Stderr, err)
		return
	}
	fixed, err := fix.Fix(parse.Source{Name: name, Code: string(code)}, fix.Opts{MigrateLambda: *lambda})
	if err != nil {
		diag.ShowError(os.Stderr, err)
		return
	}
	if s, ok := w.(io.Seeker); ok {
		s.Seek(0, io.SeekStart)
	}
	fmt.Fprint(w, fixed)
}
