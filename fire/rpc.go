package main

import (
	"bytes"
	"fmt"
	"github.com/shxsun/exec"
	"github.com/shxsun/filedist/fire/utils"
	"strings"
	"sync"
	"time"
)

var logs = make(map[string]*Log, 1000)
var tidx = utils.NewTruncIndex()

type State struct {
	mu        sync.Mutex
	result    chan error
	Running   bool
	Err       error
	StartTime time.Time
	EndTime   time.Time
}

func (s *State) Wait() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Running {
		s.Err = <-s.result
		s.Running = false
		s.EndTime = time.Now()
	}
	return s.Err
}

type Log struct {
	Stdout bytes.Buffer
	Stderr bytes.Buffer
	Cmd    *exec.Cmd
	State
}

type RpcServer struct {
	logs string // FIXME: no use here
}

func (rs *RpcServer) Run(r Container, w *Response) error {
	cmd := exec.Command(r.Name, r.Args...)
	cmd.Timeout = r.Timeout
	cmd.IsClean = r.Kill
	uid := tidx.New()
	var g = Log{}
	g.State.result = make(chan error)
	g.State.StartTime = time.Now()
	cmd.Stdout = &g.Stdout
	cmd.Stderr = &g.Stdout
	g.Cmd = cmd
	logs[uid] = &g
	err := cmd.Start()
	g.Running = true
	go func() {
		g.result <- g.Cmd.Wait()
	}()
	go g.State.Wait()
	if err != nil {
		fmt.Println(err)
		return err
	}
	w.Uid = uid
	return nil
}

func (rs *RpcServer) Kill(id string, w *Response) error {
	uid, err := tidx.Get(id)
	if err != nil {
		return err
	}
	g := logs[uid]
	return g.Cmd.KillAll()
}

func (rs *RpcServer) Wait(id string, w *Response) error {
	uid, err := tidx.Get(id)
	if err != nil {
		return err
	}
	g := logs[uid]
	err = g.Wait()

	if err != nil {
		errmsg := err.Error()
		fmt.Println(err)
		if strings.HasPrefix(errmsg, "exit status") {
			fmt.Sscan(strings.TrimLeft(errmsg, "exit status"), &w.Code)
		} else {
			w.Code = 128
		}
	}
	w.Stdout = g.Stdout.Bytes()
	return nil
}

func (rs *RpcServer) Ps(ids []string, out *[]PsResult) error {
	// no input, ouput all
	// FIXME: result not sorted
	if len(ids) == 0 {
		for uid, g := range logs {
			pr := PsResult{}
			pr.Name = g.Cmd.Args[0]
			pr.Uid = uid
			pr.Running = g.Running
			pr.StartTime = g.State.StartTime
			pr.EndTime = g.State.EndTime
			*out = append(*out, pr)
		}
		return nil
	}

	vis := make(map[string]bool)
	*out = make([]PsResult, 0)
	for _, id := range ids {
		uid, err := tidx.Get(id)
		if err != nil {
			return err
		}
		if vis[uid] == true {
			continue
		}
		vis[uid] = true
		g := logs[uid]
		pr := PsResult{}
		pr.Name = g.Cmd.Args[0]
		pr.Uid = uid
		pr.Running = g.Running
		pr.StartTime = g.State.StartTime
		pr.EndTime = g.State.EndTime
		*out = append(*out, pr)
	}
	return nil
}
