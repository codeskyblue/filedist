package main

import (
	"fmt"
	"github.com/shxsun/flags"
	"net/rpc"
)

var opts struct {
	Daemon  bool   `short:"d" long:"daemon" description:"run as server"`
	Timeout string `short:"t" long:"timeout" description:"time out [s|m|h]" default:"0s"`
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		return
	}
	// FIXME: daemon
	if opts.Daemon {

	}

	if len(args) == 0 {
		fmt.Println("Use --help for more help")
		return
	}

	// FIXME: net dial timeout

	// FIXME: call rpc service
	srv := new(RpcServer)
	rpc.Register(srv)
	rpc.HandleHTTP()
}
