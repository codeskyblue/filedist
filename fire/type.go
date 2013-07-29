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
	Uid    string
	Msg    string
	Stdout []byte
	Code   int
}

type PsResult struct {
	Uid       string
	Name      string
	StartTime time.Time
	EndTime   time.Time
	Running   bool
}
