package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/elves/upgrade-scripts-for-0.17/fix"
	"src.elv.sh/pkg/parse"
)

func main() {
	code, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	fixed, err := fix.Fix(parse.Source{Name: "[stdin]", Code: string(code)})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(fixed)
	// TODO: Support rewriting files
}
