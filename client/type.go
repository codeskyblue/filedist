package main

import (
	"time"
)

type Container struct {
	Name    string
	Args    []string
	Timeout time.Duration
	Kill    bool
}

type Response struct {
	Uid  int
	Msg  string
	Code int
}
