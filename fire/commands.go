package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"
)

func rpcCall(serviceMethod string, args interface{}, reply interface{}) (err error) {
	server := fmt.Sprintf("%s:%d", frun.Host, fsrv.Port)
	dt, err := time.ParseDuration(frun.DialTimeout) // FIXME: should this be put some thing else
	if err != nil {
		return
	}

	conn, err := net.DialTimeout("tcp", server, dt)
	if err != nil {
		return
	}
	defer conn.Close()
	io.WriteString(conn, "CONNECT "+rpc.DefaultRPCPath+" HTTP/1.0\n\n")
	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	const connected = "200 Connected to Go RPC"
	if err == nil && resp.Status == connected {
		client := rpc.NewClient(conn)
		defer client.Close()
		return client.Call(serviceMethod, args, reply)
	}
	if err == nil {
		return fmt.Errorf("unexpected HTTP response: %d", resp.Status)
	}
	return
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
	fmt.Printf("%-20s %-10s %-10s %s\n", "ID", "TIME", "CMD", "RUNNING")
	for _, pr := range out {
		fmt.Printf("%-20s %-10s %-10s %v\n", pr.Uid, pr.StartTime.Format("15:04:05"), pr.Name, pr.Running)
	}
}

func cmdKill(id string) {
	err := rpcCall("RpcServer.Kill", id, &Response{})
	if err != nil {
		log.Fatal(err)
	}
}

func cmdHelp() {
	fmt.Println("Use --help for more help")
}
