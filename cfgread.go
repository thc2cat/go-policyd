package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const cfgfile = "/usr/local/etc/polka.cf"

var cfg map[string]string

// InitCfg read polka.cf
func InitCfg() {
	cfg = make(map[string]string)
	f, err := os.Open(cfgfile)
	if err != nil {
		panic(fmt.Sprintf("Unable to read configuration file %s", cfgfile))
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
		cfg[vv[0]] = vv[1]
	}
}
