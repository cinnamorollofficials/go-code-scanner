//go:build ignore

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cinnamorollofficials/go-code-scanner/release"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "usage: archive_release <binary> <archive> <timestamp>")
		os.Exit(2)
	}
	stamp, err := time.Parse(time.RFC3339, os.Args[3])
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse archive timestamp:", err)
		os.Exit(2)
	}
	if err := release.ArchiveBinary(os.Args[1], os.Args[2], stamp); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
