// Command fsnotify provides example usage of the fsnotify library.
package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var usage = `

fsnotify is a Go library to provide cross-platform file system notifications.
This command serves as an example and debugging tool.

https://github.com/fsnotify/fsnotify

Commands:

    watch [paths]  Watch the paths for changes and print the events.
    file  [file]   Watch a single file for changes.
    dedup [paths]  Watch the paths for changes, suppressing duplicate events.
`[1:]

func exit(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, filepath.Base(os.Args[0])+": "+format+"\n", a...)
	fmt.Print("\n" + usage)
	os.Exit(1)
}

func help() {
	fmt.Printf("%s [command] [arguments]\n\n", filepath.Base(os.Args[0]))
	fmt.Print(usage)
	os.Exit(0)
}

// Print line prefixed with the time (a bit shorter than log.Print; we don't
// really need the date and ms is useful here).
func printTime(s string, args ...interface{}) {
	// fmt.Printf()
	// log.Printf(time.Now().Format("15:04:05.0000")+" "+s+"\n", args...)
	log.Printf(s+"\n", args...)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if len(os.Args) == 1 {
		help()
	}
	// Always show help if -h[elp] appears anywhere before we do anything else.
	for _, f := range os.Args[1:] {
		switch f {
		case "help", "-h", "-help", "--help":
			help()
		}
	}
	// pro()
	server := http.NewServeMux()
	server.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe("0.0.0.0:6061", server)

	cmd, args := os.Args[1], os.Args[2:]
	switch cmd {
	default:
		exit("unknown command: %q", cmd)
	case "watch":
		watch(args...)
	case "file":
		file(args...)
	case "dedup":
		dedup(args...)
	}

}
