package main

import (
	"fmt"
	"github.com/shxsun/flags"
	"github.com/shxsun/heartbeat"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strings"
	"time"
)

var fsrv struct {
	Daemon     bool     `short:"d" long:"daemon" description:"run as server" default:"false"`
	Port       int      `short:"p" long:"port" description:"port to connect or serve" default:"8119"`
	FileServer string   `long:"fs" description:"open a http file server" default:"/tmp/:/home/"`
	Expire     string   `long:"expire" description:"background job keep time after finished running, at least 10min" default:"24h"`
	Allow      []string `long:"allow" description:"allow which client can connect server"`                            // FIXME
	Unsafe     bool     `long:"unsafe" description:"allow remove client use root to execute command" default:"false"` // FIXME
	HeartBeat  string   `long:"beat" description:"open heart beat(UDP)" default:""`
}

var frun struct {
	Host        string   `short:"H" long:"host" description:"host to connect" default:"localhost"`
	User        string   `short:"u" long:"user" description:"specify which user to run"` // FIXME
	Timeout     string   `short:"t" long:"timeout" description:"time out [s|m|h]" default:"0s"`
	Background  bool     `short:"b" long:"background" description:"run in background"`
	Env         []string `short:"e" long:"env" description:"add env to runner,multi support. eg -e PATH=/bin -e TMPDIR=/tmp"` // FIXME
	DialTimeout string   `long:"dialtimeout" description:"dial timeout,unit seconds" default:"2s"`
}
var ftype struct {
	Type string `short:"m" long:"type" description:"type [run|ps|wait|kill]" default:"run" 	`
}

func startDaemon() {
	srv := NewRpcServer()
	rpc.Register(srv)
	rpc.HandleHTTP()
	if fsrv.FileServer != "" {
		for _, path := range strings.Split(fsrv.FileServer, ":") {
			http.Handle(path, http.StripPrefix(path, http.FileServer(http.Dir(path))))
		}
	}
	if fsrv.HeartBeat != "" {
		heartbeat.GoBeat(fsrv.HeartBeat)
	}
	l, e := net.Listen("tcp", fmt.Sprintf(":%d", fsrv.Port))
	if e != nil {
		log.Fatal(e)
	}
	http.Serve(l, nil)
}

func main() {
	f := flags.NewParser(nil, flags.Default)
	f.Usage = "[OPTIONS] args ..."
	f.AddGroup("Type of run", &ftype).
		AddGroup("Run", &frun).
		AddGroup("Serve", &fsrv)

	args, err := f.Parse()
	if err != nil {
		return
	}

	if fsrv.Daemon {
		startDaemon()
		return
	}

	tmo, err := time.ParseDuration(frun.Timeout)
	if err != nil {
		log.Fatal(err)
	}

	switch ftype.Type {
	case "run":
		if len(args) == 0 {
			fmt.Println("Use --help for more help")
			return
		}
		container := Container{}
		container.Args = args[1:]
		container.Name = args[0]
		container.Timeout = tmo
		container.Kill = true // FIXME: add opts
		r := Response{}
		cmdRun(container, &r)
	case "wait":
		if len(args) != 1 {
			log.Fatal("wait need one argument")
		}
		r := Response{}
		cmdWait(args[0], &r)
	case "ps":
		r := make([]PsResult, 0)
		cmdPs(args, r)
	case "kill":
		if len(args) != 1 {
			cmdHelp()
			return
		}
		cmdKill(args[0])
	default:
		log.Fatal("need specify a type. use --help for more help")
	}
}
