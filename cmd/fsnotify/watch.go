package main

import "github.com/fsnotify/fsnotify"

// This is the most basic example: it prints events to the terminal as we
// receive them.
func watch(paths ...string) {
	if len(paths) < 1 {
		exit("must specify at least one path to watch")
	}

	// Create a new watcher.
	// w, err := fsnotify.NewWatcher()
	w, err := fsnotify.WatcherRecursivelyWithExclude()
	if err != nil {
		exit("creating a new watcher: %s", err)
	}
	defer w.Close()

	// Start listening for events.
	go watchLoop(w)

	// Add all paths from the commandline
	paths, err = fsnotify.GetDirNames(paths)
	if err != nil {
		exit("add init watch path err %s", err)
	}
	for _, p := range paths {
		err = w.Add(p)
		if err != nil {
			exit("%q: %s", p, err)
		}
	}

	printTime("ready; press ^C to exit")
	<-make(chan struct{}) // Block forever
}

func watchLoop(w *fsnotify.Watcher) {
	// i := 0
	for {
		select {
		// Read from Errors.
		case err, ok := <-w.Errors:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}
			printTime("ERROR: %s", err)
		// Read from Events.
		case e, ok := <-w.Events:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}
			if e.Op.String() == "IN_CREATE|IN_ISDIR" {
				w.Add(e.Name)
			}

			// Just print the event nicely aligned, and keep track how many
			// events we've seen.
			// i++
			// printTime("%3d %s", i, e)
			printTime("Op:%s Name: %s", e.Op, e.Name)
		}
	}
}
