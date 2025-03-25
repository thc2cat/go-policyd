package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	cfg         map[string]string
	inblacklist map[string]bool
	inwhitelist map[string]bool
	usesyslog   bool
	pidfile     string
)

// InitCfg read cfgfile variable
func InitCfg(s string) {
	cfg = make(map[string]string)
	inblacklist = make(map[string]bool)
	inwhitelist = make(map[string]bool)

	f, err := os.Open(filepath.Clean(s))
	if err != nil {
		panic(fmt.Sprintf("Unable to read configuration file %s", s))
	}

	rd := bufio.NewReader(f)
	for {
		cfgline, err := rd.ReadString('\n')
		if err != nil {
			break
		}
		cfgline = strings.Trim(cfgline, " \n\r")
		cfgval := strings.SplitN(cfgline, "=", 2)
		if len(cfgval) < 2 {
			continue
		}
		switch {
		case cfgval[0] == "pidfile":
			pidfile = cfgval[1]

		case cfgval[0] == "usesyslog":
			if cfgval[1] == "yes" {
				usesyslog = true
			}
		case cfgval[0] == "blacklist":
			inblacklist[cfgval[1]] = true
		case cfgval[0] == "whitelist":
			inwhitelist[cfgval[1]] = true
		default:
			cfg[cfgval[0]] = cfgval[1]
		}
	}
	if err := f.Close(); err != nil {
		log.Print(err)
	}
}
