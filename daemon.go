//go:build linux || freebsd
// +build linux freebsd

package main

import (
	"fmt"
	"log"
	"log/syslog"
	"os"
	"time"
)

var xlog *syslog.Writer // Used by mylog

// mylog is a wrapper aroung xlog syslog Writer.
func mylog(txt string, facility string) {
	if !usesyslog {
		if lerr := log.Output(1, txt); lerr != nil {
			fmt.Print("mylog " + lerr.Error())
		}
		return
	}
	// we use syslog :
	switch facility {
	case "Err":
		if err := xlog.Err(txt); err != nil {
			fmt.Print("mylog " + err.Error())
		}
	default:
		if err := xlog.Info(txt); err != nil {
			fmt.Print("mylog " + err.Error())
		}
	}
}

func initSyslog(exename string) {
	if usesyslog {
		var err error
		xlog, err = syslog.New(syslog.LOG_DAEMON|syslog.LOG_INFO, exename)
		if err == nil {
			log.SetOutput(xlog)
			log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime)) // remove timestamp
		}
	} else {
		log.SetOutput(os.Stdout)
	}
}

func writePidfile(mypidfile string) {
	if len(pidfile) == 0 { // Not defined ?
		return
	}

	err := os.WriteFile(mypidfile, []byte(fmt.Sprintf("%d", os.Getpid())), 0o600)
	if err != nil {
		if lerr := log.Output(1, "Unable to create pidfile "+mypidfile); lerr != nil {
			fmt.Print("writePidfile " + lerr.Error())
		}
		time.Sleep(20 * time.Second)
	}
}
