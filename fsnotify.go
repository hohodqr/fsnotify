// Package fsnotify provides a cross-platform interface for file system
// notifications.
//
// Currently supported systems:
//
//	Linux 2.6.32+    via inotify
//	BSD, macOS       via kqueue
//	Windows          via ReadDirectoryChangesW
//	illumos          via FEN
package fsnotify

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// Event represents a file system notification.
type Event struct {
	// Path to the file or directory.
	//
	// Paths are relative to the input; for example with Add("dir") the Name
	// will be set to "dir/file" if you create that file, but if you use
	// Add("/path/to/dir") it will be "/path/to/dir/file".
	Name string

	// File operation that triggered the event.
	//
	// This is a bitmask and some systems may send multiple operations at once.
	// Use the Event.Has() method instead of comparing with ==.
	Op Op
}

// Op describes a set of file operations.
type Op uint32

// The operations fsnotify can trigger; see the documentation on [Watcher] for a
// full description, and check them with [Event.Has].
const (
	// A new pathname was created.

	Create Op = 1 << iota
	// IN_ACCESS
	// IN_CREATE
	// MovedTO
	// IN_MOVED_TO
	// IN_DELETE_SELF
	// IN_DELETE
	// IN_MODIFY
	// The pathname was written to; this does *not* mean the write has finished,
	// and a write can be followed by more writes.
	Write

	// The path was removed; any watches on it will be removed. Some "remove"
	// operations may trigger a Rename if the file is actually moved (for
	// example "remove to trash" is often a rename).
	Remove

	// The path was renamed to something else; any watched on it will be
	// removed.
	Rename
	// IN_MOVE_SELF
	// IN_MOVED_FROM

	// File attributes were changed.
	//
	// It's generally not recommended to take action on this event, as it may
	// get triggered very frequently by some software. For example, Spotlight
	// indexing on macOS, anti-virus software, backup software, etc.
	Chmod
	IN_ACCESS        = 0x1
	IN_ALL_EVENTS    = 0xfff
	IN_ATTRIB        = 0x4
	IN_CLASSA_HOST   = 0xffffff
	IN_CLASSA_MAX    = 0x80
	IN_CLASSA_NET    = 0xff000000
	IN_CLASSA_NSHIFT = 0x18
	IN_CLASSB_HOST   = 0xffff
	IN_CLASSB_MAX    = 0x10000
	IN_CLASSB_NET    = 0xffff0000
	IN_CLASSB_NSHIFT = 0x10
	IN_CLASSC_HOST   = 0xff
	IN_CLASSC_NET    = 0xffffff00
	IN_CLASSC_NSHIFT = 0x8
	IN_CLOSE         = 0x18
	IN_CLOSE_NOWRITE = 0x10
	IN_CLOSE_WRITE   = 0x8
	IN_CREATE        = 0x100
	IN_DELETE        = 0x200
	IN_DELETE_SELF   = 0x400
	IN_DONT_FOLLOW   = 0x2000000
	IN_EXCL_UNLINK   = 0x4000000
	IN_IGNORED       = 0x8000
	IN_ISDIR         = 0x40000000
	IN_LOOPBACKNET   = 0x7f
	IN_MASK_ADD      = 0x20000000
	IN_MASK_CREATE   = 0x10000000
	IN_MODIFY        = 0x2
	IN_MOVE          = 0xc0
	IN_MOVED_FROM    = 0x40
	IN_MOVED_TO      = 0x80
	IN_MOVE_SELF     = 0x800
	IN_ONESHOT       = 0x80000000
	IN_ONLYDIR       = 0x1000000
	IN_OPEN          = 0x20
	IN_Q_OVERFLOW    = 0x4000
)

// Common errors that can be reported.
var (
	ErrNonExistentWatch = errors.New("fsnotify: can't remove non-existent watcher")
	ErrEventOverflow    = errors.New("fsnotify: queue or buffer overflow")
	ErrClosed           = errors.New("fsnotify: watcher already closed")
)

func (o Op) String() string {
	var b strings.Builder
	if o.Has(IN_ACCESS) {
		b.WriteString("|IN_ACCESS")
	}
	if o.Has(IN_ATTRIB) {
		b.WriteString("|IN_ATTRIB")
	}
	if o.Has(IN_CLOSE) {
		b.WriteString("|IN_CLOSE")
	}
	if o.Has(IN_CLOSE_NOWRITE) {
		b.WriteString("|IN_CLOSE_NOWRITE")
	}
	if o.Has(IN_CLOSE_WRITE) {
		b.WriteString("|IN_CLOSE_WRITE")
	}
	if o.Has(IN_CREATE) {
		b.WriteString("|IN_CREATE")
	}
	if o.Has(IN_DELETE) {
		b.WriteString("|IN_DELETE")
	}
	if o.Has(IN_DELETE_SELF) {
		b.WriteString("|IN_DELETE_SELF")
	}

	if o.Has(IN_MOVED_TO) {
		b.WriteString("|IN_MOVED_TO")
	}

	if o.Has(IN_MODIFY) {
		b.WriteString("|IN_MODIFY")
	}
	if o.Has(IN_MOVE_SELF) {
		b.WriteString("|IN_MOVE_SELF")
	}
	if o.Has(IN_MOVED_FROM) {
		b.WriteString("|IN_MOVED_FROM")
	}

	if o.Has(IN_ISDIR) {
		b.WriteString("|IN_ISDIR")
	}
	if o.Has(IN_OPEN) {
		b.WriteString("|IN_OPEN")
	}
	// ----------
	if o.Has(IN_DONT_FOLLOW) {
		b.WriteString("|IN_DONT_FOLLOW")
	}
	// --------
	// if o.Has(Create) {
	// 	b.WriteString("|CREATE")
	// }
	// if o.Has(Remove) {
	// 	b.WriteString("|REMOVE")
	// }
	// if o.Has(Write) {
	// 	b.WriteString("|WRITE")
	// }
	// if o.Has(Rename) {
	// 	b.WriteString("|RENAME")
	// }
	// if o.Has(Chmod) {
	// 	b.WriteString("|CHMOD")
	// }
	if b.Len() == 0 {
		return "[no events]"
	}
	return b.String()[1:]
}

// Has reports if this operation has the given operation.
func (o Op) Has(h Op) bool { return o&h == h }

// Has reports if this event has the given operation.
func (e Event) Has(op Op) bool { return e.Op.Has(op) }

// String returns a string representation of the event with their path.
func (e Event) String() string {
	// return fmt.Sprintf("%-13s %q %+v", e.Op.String(), e.Name, e.Op)
	return fmt.Sprintf("%s %s ", e.Op, e.Name)
}

type (
	addOpt   func(opt *withOpts)
	withOpts struct {
		bufsize int
	}
)

var defaultOpts = withOpts{
	bufsize: 65536, // 64K
}

func getOptions(opts ...addOpt) withOpts {
	with := defaultOpts
	for _, o := range opts {
		o(&with)
	}
	return with
}

// WithBufferSize sets the buffer size for the Windows backend. This is a no-op
// for other backends.
//
// The default value is 64K (65536 bytes) which is the highest value that works
// on all filesystems and should be enough for most applications, but if you
// have a large burst of events it may not be enough. You can increase it if
// you're hitting "queue or buffer overflow" errors ([ErrEventOverflow]).
func WithBufferSize(bytes int) addOpt {
	return func(opt *withOpts) { opt.bufsize = bytes }
}

// Check if this path is recursive (ends with "/..." or "\..."), and return the
// path with the /... stripped.
func recursivePath(path string) (string, bool) {
	if filepath.Base(path) == "..." {
		return filepath.Dir(path), true
	}
	return path, false
}
