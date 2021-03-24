// +build linux

package main

import (
	"fmt"
	"log"
	"log/syslog"
	"os"
	"time"
)

var xlog *syslog.Writer

func initSyslog(exe string) {
	var e error
	xlog, e = syslog.New(syslog.LOG_DAEMON|syslog.LOG_INFO, exe)
	if e == nil {
		log.SetOutput(xlog)
		log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime)) // remove timestamp
	}
}

func writePidfile(pidfile string) {
	err := os.WriteFile(pidfile, []byte(fmt.Sprintf("%d", os.Getpid())), 0664)
	if err != nil {
		log.Output(1, "Unable to create pidfile "+pidfile)
		time.Sleep(20 * time.Second)
	}
}
