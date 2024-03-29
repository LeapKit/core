package assets

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// manager watches the input folder and copies all files to the output folder.
// It also watches for changes in the input folder and copies the files again.
func (m *manager) Watch() {
	err := m.CopyAll()
	if err != nil {
		log.Println(err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(fmt.Errorf("error creating watcher: %w", err))
	}

	// Add all folders within the assets folder to the watcher.
	err = filepath.Walk(m.inputFolder, func(path string, info os.FileInfo, err error) error {
		return watcher.Add(path)
	})

	if err != nil {
		panic(fmt.Errorf("error adding files to watcher: %w", err))
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					continue
				}

				needsCopy := event.Has(fsnotify.Create) || event.Has(fsnotify.Write) || event.Has(fsnotify.Rename)
				if !needsCopy {
					continue
				}

				err := m.CopyAll()
				if err != nil {
					log.Println(err)
				}

				if event.Has(fsnotify.Create) {
					watcher.Add(event.Name)
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}

				log.Println("error:", err)
			}
		}
	}()

	<-make(chan struct{})
}

// CopyAll copies all files from the input folder to the output folder.
func (m *manager) CopyAll() error {

	// Copy all files files
	err := filepath.Walk(m.inputFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Get the relative path of the file
		relativePath, err := filepath.Rel(m.inputFolder, path)
		if err != nil {
			return err
		}

		// Create the destination folder if it doesn't exist
		destFolder := filepath.Join(m.outputFolder, filepath.Dir(relativePath))
		err = os.MkdirAll(destFolder, os.ModePerm)
		if err != nil {
			return err
		}

		// Copy the file to the destination folder
		destPath := filepath.Join(destFolder, filepath.Base(relativePath))
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		destFile, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, srcFile)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error copying files: %w", err)
	}

	return nil
}
