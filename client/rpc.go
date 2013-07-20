package main

import (
	"bytes"
	"fmt"
	"github.com/shxsun/exec"
	"math/rand"
	"strings"
	"sync"
)

var logs = make(map[string]*Log, 1000)

type State struct {
	mu      sync.Mutex
	result  chan error
	Running bool
	Err     error
}

func (s *State) Wait() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Running {
		s.Err = <-s.result
		s.Running = false
	}
	return s.Err
}

type Log struct {
	Stdout bytes.Buffer
	Stderr bytes.Buffer
	Cmd    *exec.Cmd
	State
}

type RpcServer int

func (rs *RpcServer) Run(r Container, w *Response) error {
	cmd := exec.Command(r.Name, r.Args...)
	cmd.Timeout = r.Timeout
	cmd.IsClean = r.Kill
	uid := fmt.Sprintf("%d", rand.Int())
	var l = Log{}
	l.State.result = make(chan error)
	cmd.Stdout = &l.Stdout
	cmd.Stderr = &l.Stdout
	l.Cmd = cmd
	logs[uid] = &l
	err := cmd.Start()
	l.Running = true
	go func() {
		l.result <- l.Cmd.Wait()
	}()
	if err != nil {
		fmt.Println(err)
		return err
	}
	w.Uid = uid
	return nil
}

func (rs *RpcServer) Wait(id string, w *Response) error {
	l, ok := logs[id]
	if !ok {
		return fmt.Errorf("log id:(%s) not exists", id)
	}
	err := l.Wait()

	if err != nil {
		errmsg := err.Error()
		fmt.Println(err)
		if strings.HasPrefix(errmsg, "exit status") {
			fmt.Sscan(strings.TrimLeft(errmsg, "exit status"), &w.Code)
		} else {
			w.Code = 128
		}
	}
	w.Stdout = l.Stdout.Bytes()
	return nil
}

func (rs *RpcServer) Ps(id []string, out *[]PsResult) error {
	// FIXME: id not used at present
	*out = make([]PsResult, 0)
	for uid, l := range logs {
		pr := PsResult{}
		pr.Name = l.Cmd.Args[0]
		pr.Uid = uid
		pr.Running = l.Running
		*out = append(*out, pr)
	}
	return nil
}
