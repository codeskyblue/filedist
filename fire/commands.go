package main

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
)

// FIXME: net dial timeout
func rpcCall(serviceMethod string, args interface{}, reply interface{}) error {
	client, err := rpc.DialHTTP("tcp", fmt.Sprintf("%s:%d", frun.Host, fsrv.Port))
	defer client.Close()
	if err != nil {
		log.Fatal(err)
	}
	return client.Call(serviceMethod, args, reply)
}

func cmdRun(container Container, r *Response) {
	err := rpcCall("RpcServer.Run", container, r)
	if err != nil {
		log.Fatal(err)
	}
	if frun.Background {
		fmt.Println(r.Uid)
		return
	}
	cmdWait(r.Uid, r)
}

func cmdWait(id string, r *Response) {
	err := rpcCall("RpcServer.Wait", id, r)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(string(r.Stdout))
	os.Exit(r.Code)
}

func cmdPs(id []string, out []PsResult) {
	err := rpcCall("RpcServer.Ps", id, &out)
	if err != nil {
		log.Fatal(err)
	}
	for idx, pr := range out {
		fmt.Printf("%-3d %-10s %-10s running:%v\n", idx, pr.Uid, pr.Name, pr.Running)
	}
}
