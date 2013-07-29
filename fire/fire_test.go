package main

import (
	"testing"
)

const PORT = 22222

func init() {
	fsrv.Port = PORT
	go startDaemon()
}

func TestNormalRun(t *testing.T) {
	c := Container{}
	c.Name = "echo"
	c.Args = []string{"-n", "hello"}
	c.Timeout = 0
	frun.DialTimeout = "0s"
	r := &Response{}
	cmdRun(c, r)
}
