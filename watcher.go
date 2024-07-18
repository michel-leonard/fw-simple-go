package watcher

import (
	"fw/pkg/config"
	"fw/pkg/processor"
	"github.com/fsnotify/fsnotify"
	"log"
	"sync"
	"time"
)

// Watch a file and calls processor routines when it changes.
// A delay of 100 ms after the last modification is used to avoid confusion.
func WatchFile(config config.Config, path string, wg *sync.WaitGroup) {
	defer wg.Done()
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer func(watcher *fsnotify.Watcher) {
		_ = watcher.Close()
	}(watcher)
	err = watcher.Add(path)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Watching the file '%s'.\n", path)
	var offset int64
	var timer *time.Timer
	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				if timer != nil {
					timer.Stop()
				}
				timer = time.NewTimer(100 * time.Millisecond)
				go func() {
					for range timer.C {
						processor.ProcessFileChange(config, path, &offset)
					}
				}()
			}
		case err := <-watcher.Errors:
			log.Println("error:", err)
		}
	}
}
