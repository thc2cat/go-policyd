//+build windows

// Windows fake functions when editing on windows
// -- this is intended to work on Unix first.
package main

type myXlog int

var xlog myXlog

func (myXlog) Err(string)   {}
func (myXlog) Info(string)  {}
func initSyslog(s string)   {}
func writePidfile(s string) {}
