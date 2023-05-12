package fsnotify

import (
	"log"
	"os"
	"path/filepath"
)

func visit(files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		*files = append(*files, path)
		return nil
	}
}

func GetDirNames(names []string) ([]string, error) {
	var files []string

	for _, v := range names {
		err := filepath.Walk(v, visit(&files))
		if err != nil {
			return nil, err
		}
	}

	return files, nil

}
