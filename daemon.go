// +build linux

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"log/syslog"
	"os"
	"os/exec"
	"syscall"
	"time"
)

var (
	xlog *syslog.Writer
)

func initSyslog(exe string) {
	e error
	xlog, e = syslog.New(syslog.LOG_DAEMON|syslog.LOG_INFO, exe)
	if e == nil {
		log.SetOutput(xlog)
		log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime)) // remove timestamp
	}
}

// See https://github.com/golang/go/issues/227#issuecomment-235996646
func daemon(nochdir, noclose int) {
	child := os.Getenv("XCHILDX")

	if child == "" {
		// I am the parent, spawn child to run as daemon
		binary, err := exec.LookPath(os.Args[0])
		if err != nil {
			log.Fatalln("Failed to lookup binary:", err)
			os.Exit(1)
		}
		newenv := []string{"XCHILDX=1"}
		_, err = os.StartProcess(binary, os.Args, &os.ProcAttr{Dir: "", Env: newenv,
			Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}, Sys: nil})
		if err != nil {
			log.Fatalln("Failed to start process:", err)
			os.Exit(1)
		}
		os.Exit(0)
	} else {
		// I am the child, i.e. the daemon, start new session and detach from terminal
		_, err := syscall.Setsid()
		if err != nil {
			log.Fatalln("Failed to create new session:", err)
		}
		writePidfile("/var/run/policyd.pid")
		if nochdir == 0 {
			os.Chdir("/")
		}
		if noclose == 0 {
			file, err := os.OpenFile("/dev/null", os.O_RDWR, 0)
			if err != nil {
				log.Fatalln("Failed to open /dev/null:", err)
			}
			syscall.Dup2(int(file.Fd()), int(os.Stdin.Fd()))
			syscall.Dup2(int(file.Fd()), int(os.Stdout.Fd()))
			syscall.Dup2(int(file.Fd()), int(os.Stderr.Fd()))
			file.Close()
		}
	}
}

func writePidfile(pidfile string) {
	err := ioutil.WriteFile(pidfile, []byte(fmt.Sprintf("%d", os.Getpid())), 0664)
	if err != nil {
		log.Output(1, "Unable to create pidfile "+pidfile)
		time.Sleep(20 * time.Second)
	}
}
