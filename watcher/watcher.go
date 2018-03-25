package watcher

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/fsnotify/fsnotify"
)

type CallBack func(fileName string)

type FileInfo struct {
	isDir bool
	// mTime       time.Time
	// provisioned bool
}

type Watcher struct {
	root     string
	fileMap  map[string]*FileInfo
	watcher  *fsnotify.Watcher
	indexCB  CallBack
	removeCB CallBack
}

func NewWatcher(root string, indexCB CallBack, removeCB CallBack, skipIndexing bool) *Watcher {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	watcher := &Watcher{root: root, fileMap: map[string]*FileInfo{}, watcher: w, indexCB: indexCB, removeCB: removeCB}
	watcher.tweakLimit()
	if skipIndexing == false {
		watcher.addDir(root)
	}
	return watcher
}

func (watcher *Watcher) Start() {
	defer watcher.watcher.Close()
	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.watcher.Events:
				log.Infoln("event:", event)
				if strings.HasPrefix(event.Name, ".") || event.Name == "" || event.Name == " " {
					continue
				}
				if watcher.isDir(event.Name) {
					if (event.Op&fsnotify.Remove == fsnotify.Remove) ||
						(event.Op&fsnotify.Rename == fsnotify.Rename) {
						watcher.removeDir(event.Name)
					} else if (event.Op&fsnotify.Write == fsnotify.Write) ||
						(event.Op&fsnotify.Create == fsnotify.Create) {
						watcher.addDir(event.Name)
					}
				} else {
					if (event.Op&fsnotify.Remove == fsnotify.Remove) ||
						(event.Op&fsnotify.Rename == fsnotify.Rename) {
						watcher.removeFile(event.Name)
					} else if (event.Op&fsnotify.Write == fsnotify.Write) ||
						(event.Op&fsnotify.Create == fsnotify.Create) {
						watcher.indexFile(event.Name)
					}
				}
			case err := <-watcher.watcher.Errors:
				fmt.Println("error:", err)
			}
		}
	}()
	<-done
}

func (watcher *Watcher) indexFile(fileName string) {
	fileInfo, ok := watcher.fileMap[fileName]
	if !ok {
		fileInfo = &FileInfo{isDir: false}
		watcher.fileMap[fileName] = fileInfo
	}
	log.Debugln("indexFile: Indexing", fileName)
	watcher.indexCB(fileName)
}

func (watcher *Watcher) removeFile(fileName string) {
	if _, ok := watcher.fileMap[fileName]; !ok {
		log.Warnln("removeFile: file not in map", fileName)
		return
	}
	log.Debugln("removeFile: Removing", fileName)
	watcher.removeCB(fileName)
	delete(watcher.fileMap, fileName)
}

func (watcher *Watcher) addDir(dirName string) {
	if fileInfo, ok := watcher.fileMap[dirName]; !ok {
		fileInfo = &FileInfo{isDir: true}
		watcher.fileMap[dirName] = fileInfo
	}
	watcher.watcher.Add(dirName)
	// Process all files under dir
	for fileName, fileInfo := range watcher.dirWalk(dirName) {
		if fileInfo.isDir {
			watcher.watcher.Add(fileName)
			watcher.fileMap[fileName] = fileInfo
		} else {
			watcher.indexFile(fileName)
		}
	}
}

func (watcher *Watcher) removeDir(dirName string) {
	if _, ok := watcher.fileMap[dirName]; !ok {
		log.Warnln("removeDir: dir not in map", dirName)
		return
	}
	// Process all files under dir
	for fileName, fileInfo := range watcher.mapWalk(dirName) {
		if fileInfo.isDir {
			delete(watcher.fileMap, fileName)
		} else {
			watcher.removeFile(fileName)
		}
	}
}

func (watcher *Watcher) mapWalk(root string) map[string]*FileInfo {
	fileMap := map[string]*FileInfo{}
	for fileName, fileInfo := range watcher.fileMap {
		if strings.HasPrefix(fileName, root) {
			fileMap[fileName] = fileInfo
		}
	}
	return fileMap
}

func (watcher *Watcher) dirWalk(root string) map[string]*FileInfo {
	l := map[string]*FileInfo{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Warnf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		l[path] = &FileInfo{isDir: info.IsDir()}
		return nil
	})
	if err != nil {
		return nil
	}
	return l
}

func (watcher *Watcher) isDir(filename string) bool {
	if f, ok := watcher.fileMap[filename]; ok {
		return f.isDir
	}
	fi, err := os.Stat(filename)
	if err != nil {
		if strings.HasSuffix(err.Error(), "no such file or directory") {
			return false
		}
		log.Warnln("isDir: os.Stat error: ", err)
		return false
	}
	return fi.Mode().IsDir()
}

func (watcher *Watcher) tweakLimit() {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		log.Errorln("Unable to obtain rLimit", err)
	}
	if rLimit.Cur < rLimit.Max {
		rLimit.Max = 64000
		rLimit.Cur = 64000
		err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
		if err != nil {
			log.Warnln("Unable to increase number of open files limit", err)
		}
	}
}
