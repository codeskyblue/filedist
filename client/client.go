package main

import (
	"bytes"
	"fmt"
	"github.com/shxsun/exec"
	"math/rand"
	"net/rpc"
	"time"
)

type Container struct {
	Name    string
	Args    []string
	Timeout time.Duration
	Kill    bool
}

var logs = make(map[string]*Log, 1000)

type Log struct {
	Stdout bytes.Buffer
	Stderr bytes.Buffer
	Cmd    *exec.Cmd
}

type RpcServer int

type Response struct {
	Uid  int
	Msg  string
	Code int
}

func (rs *RpcServer) Run(r Container, w *string) error {
	cmd := exec.Command(r.Name, r.Args...)
	cmd.Timeout = r.Timeout
	cmd.IsClean = r.Kill
	uid := fmt.Sprintf("%d", rand.Int())
	var l = Log{}
	logs[uid] = &l
	l.Cmd = cmd
	cmd.Stdout = &l.Stdout
	cmd.Stderr = &l.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func main() {
	srv := new(RpcServer)
	rpc.Register(srv)
	rpc.HandleHTTP()
	fmt.Println("Client")
}
