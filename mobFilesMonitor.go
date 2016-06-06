package main

import (
	"sync/atomic"
	"github.com/guotie/deferinit"
	"sync"
	"github.com/smtc/glog"
	"github.com/howeyc/fsnotify"
	"strings"
)

type counter struct {
	val int32
}
func (c *counter) increment() {
	atomic.AddInt32(&c.val, 1)
}

func init() {
	deferinit.AddRoutine(mobFileProcess)
}

/**
号码包同步
 */
func mobFileProcess(ch chan struct{},wg *sync.WaitGroup)  {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		glog.Error("mobFileProcess: fsnotify newWatcher is error! err: %s \n", err.Error())
		return
	}
	var  modifyReceived counter
	done := make(chan bool)
	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				glog.Info("mobFileProcess: fsnotify watcher fileName: %s is change!  ev: %v \n", ev.Name, ev)
				if ev.IsModify()&&strings.Index(ev.Name,"zfb_czhd_")>=0 {
					modifyReceived.increment()
					if modifyReceived.val % 2 == 0 {
						go func(filePath string) {
							fileLoad(ev.Name)
						}(ev.Name)
					}
				}
			case err := <-watcher.Error:
				glog.Error("mobFileProcess: fsnotify watcher is error! err: %s \n", err.Error())
			}
		}
		done <- true
	}()
	err = watcher.WatchFlags(mobFilesPath,fsnotify.FSN_MODIFY)
	if err != nil {
		glog.Error("mobFileProcess watch error. mobFiles: %s  err: %s \n", mobFilesPath, err.Error())
	}

	// Hang so program doesn't exit
	<-ch

	/* ... do stuff ... */
	watcher.Close()
	wg.Done()
}