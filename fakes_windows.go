//+build windows

// Windows fake functions -- polka is intended to work on Unix.
package main

type myXlog int

var xlog myXlog

func (myXlog) Err(string)   {}
func (myXlog) Info(string)  {}
func initSyslog(s string)   {}
func writePidfile(s string) {}
