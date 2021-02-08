package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var cfg map[string]string
var inblacklist map[string]bool
var inwhitelist map[string]bool

// InitCfg read cfgfile variable
func InitCfg(s string) {
	cfg = make(map[string]string)
	inblacklist = make(map[string]bool)
	inwhitelist = make(map[string]bool)

	f, err := os.Open(s)
	if err != nil {
		panic(fmt.Sprintf("Unable to read configuration file %s", s))
	}
	defer f.Close()
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
		case cfgval[0] == "blacklist":
			inblacklist[cfgval[1]] = true
		case cfgval[0] == "whitelist":
			inwhitelist[cfgval[1]] = true
		default:
			cfg[cfgval[0]] = cfgval[1]
		}
	}
}
