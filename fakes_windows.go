//+build windows

// Windows fake functions -- polka is intended to work on Unix.
package main

type myXlog int

var xlog myXlog

func (myXlog) Err(string)         {}
func (myXlog) Info(string)        {}
func daemon(nochdir, noclose int) {}
func initSyslog()                 {}
