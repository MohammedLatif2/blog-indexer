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

type strList []string

type watcherArgs struct {
	watcher      *fsnotify.Watcher
	root         string
	el           *elastic.Elastic
	watchList    strList
	indexedFiles strList
}

func Watcher(root string, el *elastic.Elastic) {
	tweakLimit()
	watcherArgs := newWatcherArgs(root, el)
	dirs, files := getDirsAndFilesFrom(watcherArgs.root)
	watcherArgs.setWatcher()
	defer watcherArgs.watcher.Close()
	watcherArgs.indexFiles(files)
	watcherArgs.updateWatchList(dirs)
	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcherArgs.watcher.Events:
				// fmt.Println("event:", event)
				if watcherArgs.isDir(event.Name) {
					// if event.Op&fsnotify.Write == fsnotify.Write {
					// 	Do something?
					// }
					if event.Op&fsnotify.Create == fsnotify.Create {
						watcherArgs.addNewDir(event.Name)
					}
					if (event.Op&fsnotify.Remove == fsnotify.Remove) ||
						(event.Op&fsnotify.Rename == fsnotify.Rename) {
						watcherArgs.removeDir(event.Name)
					}
				} else if isMarkdown(event.Name) {
					if (event.Op&fsnotify.Write == fsnotify.Write) ||
						(event.Op&fsnotify.Create == fsnotify.Create) {
						watcherArgs.writeFile(event.Name)
					}
					if (event.Op&fsnotify.Remove == fsnotify.Remove) ||
						(event.Op&fsnotify.Rename == fsnotify.Rename) {
						watcherArgs.removeFile(event.Name)
					}
				}
			case err := <-watcherArgs.watcher.Errors:
				fmt.Println("error:", err)
			}
		}
	}()

	<-done
}

func newWatcherArgs(root string, el *elastic.Elastic) *watcherArgs {
	return &watcherArgs{root: root, el: el}
}

func (watcherArgs *watcherArgs) setWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	watcherArgs.watcher = watcher
}

func (watcherArgs *watcherArgs) updateWatchList(watchList strList) {
	for i := 0; i < len(watchList); i++ {
		// Check for duplicate directories
		dir := watchList[i]
		if watcherArgs.watchList.findStr(dir) != -1 {
			continue
		}
		// Add directory to watchlist
		err := watcherArgs.watcher.Add(dir)
		if err != nil {
			log.Fatal(err)
		}
		watcherArgs.watchList = append(watcherArgs.watchList, dir)
		// fmt.Println("ADDED DIR TO WATCHLIST ", dir)
	}
}

func (watcherArgs *watcherArgs) indexFiles(files strList) {
	for _, file := range files {
		watcherArgs.writeFile(file)
	}
}

func (watcherArgs *watcherArgs) writeFile(name string) {
	err := watcherArgs.el.IndexDoc(name, watcherArgs.root)
	if err != nil {
		// fmt.Println("ERROR IDXING DOC ", name)
		log.Println(err)
		return
	}
	// fmt.Println("IDX DOC ", name)
	if i := watcherArgs.indexedFiles.findStr(name); i != -1 {
		return
	}
	watcherArgs.indexedFiles = append(watcherArgs.indexedFiles, name)
}

func (watcherArgs *watcherArgs) removeFile(name string) {
	i := watcherArgs.indexedFiles.findStr(name)
	if i == -1 {
		return
	}
	err := watcherArgs.el.DeleteDoc(name)
	if err != nil {
		// fmt.Println("ERROR REMOVING DOC ", name)
		log.Println(err)
		return
	}
	// fmt.Println("REMDOC: ", name)
	watcherArgs.indexedFiles = removeStrAt(i, watcherArgs.indexedFiles)
}

func (watcherArgs *watcherArgs) removeDir(name string) {
	if i := watcherArgs.watchList.findStr(name); i != -1 {
		watcherArgs.watchList = removeStrAt(i, watcherArgs.watchList)
		watcherArgs.watcher.Remove(name)
		// fmt.Println("REMOVED DIR FROM WATCHLIST:", name)
		watcherArgs.removeDirsWithPrefix(name)
		watcherArgs.removeFilesWithPrefix(name)
	}
}

func (watcherArgs *watcherArgs) removeDirsWithPrefix(prefix string) {
	removedDirs := 0
	for i := 0; i < len(watcherArgs.watchList); i++ {
		idx := i - removedDirs
		dir := watcherArgs.watchList[idx]
		if strings.HasPrefix(dir, prefix) == false {
			continue
		}
		// fmt.Println("REMOVED DIR FROM WATCHLIST:", dir)
		watcherArgs.watchList = removeStrAt(idx, watcherArgs.watchList)
		watcherArgs.watcher.Remove(dir)
		removedDirs++
	}
}

func (watcherArgs *watcherArgs) removeFilesWithPrefix(prefix string) {
	removedFiles := 0
	for i := 0; i < len(watcherArgs.indexedFiles); i++ {
		idx := i - removedFiles
		file := watcherArgs.indexedFiles[idx]
		if strings.HasPrefix(file, prefix) == false {
			continue
		}
		err := watcherArgs.el.DeleteDoc(file)
		if err != nil {
			// fmt.Println("ERROR REMOVING DOC ", name)
			log.Println(err)
			return
		}
		// fmt.Println("REMDOC: ", file)
		watcherArgs.indexedFiles = removeStrAt(idx, watcherArgs.indexedFiles)
		removedFiles++
	}
}

func (watcherArgs *watcherArgs) addNewDir(name string) {
	dirs, files := getDirsAndFilesFrom(name)
	watcherArgs.indexFiles(files)
	watcherArgs.updateWatchList(dirs)
}

func (strList strList) findStr(str string) int {
	for i, val := range strList {
		if val == str {
			return i
		}
	}
	return -1
}

func getDirsAndFilesFrom(root string) (strList, strList) {
	files := strList{}
	dirs := strList{}
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

func (watcherArgs *watcherArgs) isDir(filename string) bool {
	if watcherArgs.watchList.findStr(filename) != -1 {
		return true
	}
	fi, err := os.Stat(filename)
	if err != nil {
		if strings.HasSuffix(err.Error(), "no such file or directory") {
			return false
		}
		fmt.Println(err)
		return false
	}
	return fi.Mode().IsDir()
}

func isMarkdown(filename string) bool {
	return filepath.Ext(filename) == ".md"
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

func removeStrAt(idx int, list strList) strList {
	return append(list[:idx], list[idx+1:]...)
}
