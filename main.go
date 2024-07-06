package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

// GetFiles returns a list of files in the given directory
func GetFiles(root string, fileTypes []string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && slices.Contains(fileTypes, filepath.Ext(path)) {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

// Watcher watches a single file for modifications
func Watcher(filePath string, fileChanges chan string, wg *sync.WaitGroup) {
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
		time.Sleep(300 * time.Microsecond)
	}
}

// ExecuteCommand executes a shell command
func ExecuteCommand(command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	var (
		path    string
		types   string
		commnad string
		verbose bool
	)

	flag.StringVar(&path, "path", "./", "Specify the directory path")
	flag.StringVar(&types, "types", ".go", "Specify file types to watch")
	flag.StringVar(&commnad, "commnad", "go run main.go", "Specify the command to execute when a file changes")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose mode")

	flag.Parse()

	files, err := GetFiles(path, strings.Split(types, ","))
	if err != nil {
		log.Fatalln("There is an error in getting the list of files: ", err)
	}

	fileChanges := make(chan string)
	wg := sync.WaitGroup{}

	// Start watcher for each file
	for _, file := range files {
		wg.Add(1)
		go Watcher(file, fileChanges, &wg)
	}

	go func() {
		for filePath := range fileChanges {
			if verbose {
				fmt.Println("file changed: ", filePath)
			}
			ExecuteCommand(commnad)
		}
	}()

	wg.Wait()
	close(fileChanges)
}
