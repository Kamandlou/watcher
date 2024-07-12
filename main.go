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

const (
	MODIFICATION uint8 = 1
	FSNOTIFY     uint8 = 2
)

var mode uint8

// Define flags variables
var (
	path    string
	types   string
	command string
	verbose bool
	period  uint64
	delay   uint64
)

var files []string

func main() {
	flag.StringVar(&path, "path", "./", "Specify the directory path")
	flag.StringVar(&types, "types", "", "Specify file types to watch")
	flag.StringVar(&command, "command", "", "Specify the command to execute when a file changes")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose mode")
	flag.Uint64Var(&period, "period", 0, "Set period time to watch")
	flag.Uint64Var(&delay, "delay", 0, "Set delay time to execute command")

	flag.Parse()

	mode = FSNOTIFY

	if period > 0 {
		mode = MODIFICATION
	}

	pathInfo, err := os.Stat(path)

	if err != nil {
		log.Fatalln("Path is incorrect: ", err)
	}

	if pathInfo.IsDir() {
		var fileTypes []string
		if types == "" {
			fileTypes = nil
		} else {
			fileTypes = strings.Split(types, ",")
		}

		files, err = GetDirectoryFiles(path, fileTypes)

		if err != nil {
			log.Fatalln("There is an error in getting the list of files: ", err)
		}
	} else {
		files = append(files, path)
	}

	wg := sync.WaitGroup{}

	switch mode {
	case FSNOTIFY:
		InitFsnotifyMode(&wg)
		break
	case MODIFICATION:
		InitModificationMode(&wg)
	}

	wg.Wait()
}

// GetDirectoryFiles Returns a list of files in the given directory with types filter
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
				}
			} else {
				files = append(files, path)
			}
		}
		return nil
	})
	return files, err
}

// FileWatcher Watches a single file for modifications
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
					if verbose {
						Logger(event.Name)
					}
					go func() {
						time.Sleep(time.Duration(delay) * time.Millisecond)
						ExecuteCommand(command)
					}()
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
			go func() {
				time.Sleep(time.Duration(delay) * time.Millisecond)
				ExecuteCommand(command)
			}()
		}
	}()
}
