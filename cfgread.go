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
		lin, err := rd.ReadString('\n')
		if err != nil {
			break
		}
		lin = strings.Trim(lin, " \n\r")
		vv := strings.SplitN(lin, "=", 2)
		if len(vv) < 2 {
			continue
		}
		switch {
		case vv[0] == "blacklist":
			inblacklist[vv[1]] = true
		case vv[0] == "whitelist":
			inwhitelist[vv[1]] = true
		default:
			cfg[vv[0]] = vv[1]
		}
	}
}
