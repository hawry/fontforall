// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/exp/winfsnotify"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
)

var elog debug.Log

type myservice struct{}

func loadNewFont(filepath string) {
	log.Printf("adding '%s' to user font space\n", filepath)
	//sleep to wait for file to be completely written
	time.Sleep(100 * time.Millisecond)
	mod := syscall.NewLazyDLL("Gdi32.dll")
	proc := mod.NewProc("AddFontResourceW")
	ret, _, _ := proc.Call(uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(filepath))))
	if ret == 1 {
		elog.Info(10, fmt.Sprintf("added %s to user font space", filepath))
	} else {
		elog.Warning(11, fmt.Sprintf("call to fontresource returned %d for %s", ret, filepath))
	}
}

func (m *myservice) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}
	fasttick := time.Tick(500 * time.Millisecond)
	//slowtick := time.Tick(2 * time.Second) /for pause/continue
	tick := fasttick
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	publicDir := os.Getenv("PUBLIC")
	fontDir := fmt.Sprintf("%s\\%s", strings.TrimSuffix(publicDir, "\\"), "fonts")

	if _, err := os.Stat(fontDir); err != nil {
		if os.IsNotExist(err) {
			err := os.Mkdir(fontDir, os.ModePerm)
			if err != nil {
				elog.Error(8, fmt.Sprintf("could not create (or find) font directory, control the user permissions (%v)", err))
				changes <- svc.Status{State: svc.StopPending}
				return
			}
			elog.Info(22, fmt.Sprintf("created shared font dir %s", fontDir))
		}
	}

	files, err := ioutil.ReadDir(fontDir)
	if err != nil {
		var env string
		for _, e := range os.Environ() {
			env += e
		}
		elog.Error(7, fmt.Sprintf("could not read dir list (%v) (env=%+v))", err, env))
		changes <- svc.Status{State: svc.StopPending}
		return
	}
	//add existing fonts
	for _, f := range files {
		fontFile := fmt.Sprintf("%s\\%s", fontDir, f.Name())
		loadNewFont(fontFile)
	}

	watcher, err := winfsnotify.NewWatcher()
	if err != nil {
		elog.Error(2, fmt.Sprintf("could not start notify watcher (%v)", err))
		changes <- svc.Status{State: svc.StopPending}
		return
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event := <-watcher.Event:
				if event.Mask == 256 {
					go loadNewFont(event.Name)
				}
			case err := <-watcher.Error:
				if IsShutdownError(err) {
					return
				}
				elog.Error(3, fmt.Sprintf("watcher returned an error (%v)", err))
			}
		}
	}()

	err = watcher.AddWatch(fontDir, winfsnotify.FS_CREATE)
	if err != nil {
		elog.Error(4, fmt.Sprintf("watcher could not start to watch directory %s (%v)", fontDir, err))
		err = controlService(svcName, svc.Stop, svc.Stopped)
	}

SVCLOOP:
	for {
		select {
		case <-tick:
			//do nothing special here, we're already running some things
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				watcher.Error <- &ShutdownMessageError{}
				break SVCLOOP
			default:
				elog.Error(1, fmt.Sprintf("unexpected control requests #%d", c))
			}
		}
	}
	changes <- svc.Status{State: svc.StopPending}
	return
}

func runService(name string, isDebug bool) {
	var err error
	if isDebug {
		elog = debug.New(name)
	} else {
		elog, err = eventlog.Open(name)
		if err != nil {
			return
		}
	}
	defer elog.Close()

	elog.Info(1, fmt.Sprintf("starting %s service", name))
	run := svc.Run
	if isDebug {
		run = debug.Run
	}
	err = run(name, &myservice{})
	if err != nil {
		elog.Error(1, fmt.Sprintf("%s service failed: %v", name, err))
		return
	}
	elog.Info(1, fmt.Sprintf("%s service stopped", name))
}
