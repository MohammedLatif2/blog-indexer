package watcher

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/MohammedLatif2/blog-indexer/elastic"

	"github.com/fsnotify/fsnotify"
	jww "github.com/spf13/jwalterweatherman"
)

func Watcher(root string, el *elastic.Elastic) {
	tweakLimit()
	dirs, _ := getDirsAndFilesFrom(root)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	watchList := []string{}
	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					write(event.Name)
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					create(event.Name, &watchList, watcher)
				}
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					remove(event.Name, &watchList, watcher)
				}
				if event.Op&fsnotify.Rename == fsnotify.Rename {
					remove(event.Name, &watchList, watcher)
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()
	for _, dir := range dirs {
		watchList = append(watchList, dir)
		err = watcher.Add(dir)
		if err != nil {
			log.Fatal(err)
		}
	}
	<-done
}

func tweakLimit() {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		jww.ERROR.Println("Unable to obtain rLimit", err)
	}
	if rLimit.Cur < rLimit.Max {
		rLimit.Max = 64000
		rLimit.Cur = 64000
		err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
		if err != nil {
			jww.WARN.Println("Unable to increase number of open files limit", err)
		}
	}
}

func remove(name string, watchList *[]string, watcher *fsnotify.Watcher) {
	if i := inWatchList(name, *watchList); i != -1 {
		*watchList = append((*watchList)[:i], (*watchList)[i+1:]...)
		log.Println("Dir removed from watchlist:", name)
		watcher.Remove(name)
	} else if isMarkdown(name) {
		log.Println("REMOVE: ", name)
	}
}

func write(name string) {
	if isMarkdown(name) {
		log.Println("INDEX: ", name)
	}
}

func create(name string, watchList *[]string, watcher *fsnotify.Watcher) {
	if isDir(name) {
		*watchList = append(*watchList, name)
		err := watcher.Add(name)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Dir added to watchlist:", name)
	} else {
		write(name)
	}
}

func inWatchList(str string, watchList []string) int {
	for i, val := range watchList {
		if val == str {
			return i
		}
	}
	return -1
}

func getDirsAndFilesFrom(root string) ([]string, []string) {
	files := []string{}
	dirs := []string{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() {
			dirs = append(dirs, path)
			return nil
		} else if strings.HasSuffix(info.Name(), ".md") {
			files = append(files, path)
			return nil
		}
		return nil
	})
	if err != nil {
		return nil, nil
	}
	return dirs, files
}

func isDir(filename string) bool {
	fi, err := os.Stat(filename)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return fi.Mode().IsDir()
}

func isMarkdown(filename string) bool {
	return filepath.Ext(filename) == ".md"
}
