package fsnotify

import (
	"testing"
)

func TestReadDir(t *testing.T) {
	var d []string

	d, _ := GetDirNames()
	for _, v := range d {
		// t.Error(v.Size())
		t.Error(v)
		// t.Error(v.IsDir())
	}
}
