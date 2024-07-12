package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// GetDirectoryFiles returns a list of files in the given directory with types filter
func GetDirectoryFiles(root string, fileTypes []string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if len(fileTypes) > 0 {
				if slices.Contains(fileTypes, filepath.Ext(path)) {
					files = append(files, path)
				} else {
					return nil
				}
			}

			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// FileWatcher watches a single file for modifications
func FileWatcher(filePath string, fileChanges chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	var lastModTime time.Time
	for {
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				log.Printf("File does not exist: %s", filePath)
				break
			}
			log.Printf("Error statting file: %s", err)
			break
		}

		newModTime := fileInfo.ModTime()

		if lastModTime.IsZero() {
			lastModTime = newModTime
		}

		if newModTime.After(lastModTime) {
			lastModTime = newModTime
			fileChanges <- filePath
		}

		time.Sleep(time.Duration(period) * time.Millisecond)
	}
}

// ExecuteCommand executes a shell command
func ExecuteCommand(command string) error {
	time.Sleep(time.Duration(delay) * time.Millisecond)
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Logger(filePath string) {
	log.Printf("file modified: %v", filePath)
}

func InitFsnotifyMode(wg *sync.WaitGroup) {
	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		log.Fatalln("fsnotify error: ", err)
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					Logger(event.Name)
					go ExecuteCommand(commnad)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	// Add a files to watcher.
	for _, filePath := range files {
		err = watcher.Add(filePath)
		wg.Add(1)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func InitModificationMode(wg *sync.WaitGroup) {
	fileChanges := make(chan string)

	// Start watcher for each file
	for _, file := range files {
		wg.Add(1)
		go FileWatcher(file, fileChanges, wg)
	}

	go func() {
		for filePath := range fileChanges {
			if verbose {
				Logger(filePath)
			}
			go ExecuteCommand(commnad)
		}
	}()
}

const (
	MODIFICATION uint8 = 1
	FSNOTIFY     uint8 = 2
)

var MODE uint8

var (
	path    string
	types   string
	commnad string
	verbose bool
	period  uint64
	delay   uint64
)

var files []string

func main() {
	flag.StringVar(&path, "path", "./", "Specify the directory path")
	flag.StringVar(&types, "types", "", "Specify file types to watch")
	flag.StringVar(&commnad, "command", "", "Specify the command to execute when a file changes")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose mode")
	flag.Uint64Var(&period, "period", 0, "Set period time to watch")
	flag.Uint64Var(&delay, "delay", 0, "Set delay time to execute command")

	flag.Parse()

	if period == 0 {
		MODE = FSNOTIFY
	} else {
		MODE = MODIFICATION
	}

	pathInfo, err := os.Stat(path)

	if err != nil {
		log.Fatalln("Path is incorrect: ", err)
	}

	if pathInfo.IsDir() {
		if types == "" {
			files, err = GetDirectoryFiles(path, nil)
		} else {
			files, err = GetDirectoryFiles(path, strings.Split(types, ","))
		}

		if err != nil {
			log.Fatalln("There is an error in getting the list of files: ", err)
		}

	} else {
		files = append(files, path)
	}

	wg := sync.WaitGroup{}

	switch MODE {
	case FSNOTIFY:
		InitFsnotifyMode(&wg)

	case MODIFICATION:
		InitModificationMode(&wg)
	}

	wg.Wait()
}
