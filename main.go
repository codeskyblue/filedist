/*
   This is a demo for file distribution. Like BT but no torrent file.
*/
package main

import (
	"fmt"
	"github.com/shxsun/beelog"
	"github.com/shxsun/flags"
	"io/ioutil"
	"log"
	"os/exec"
//	"path/filepath"
	"runtime"
	"strings"
	"time"
    "os"
)

var src = make(chan string)
var dst = make(chan string)
var Source = []string{}
var Dest = []string{}
var Path string
var left = -1

func initSource(S []string) {
	for _, d := range S {
		beelog.Trace("add source:", d)
		src <- d
	}
	beelog.Info("src init done")
}
func initDest(D []string) {
	left = len(D)
	log.Println("Total target: ", len(D))
	for _, d := range D {
		beelog.Trace("add dst", d)
		dst <- d
	}
	beelog.Trace("dst done")
}

// push to the todo channel
func push(ch chan string, data string) {
	go func() {
		ch <- data
	}()
}

// file copy function
func copywork(s string, d string) {
	wgetParams := []string{"wget", "-nv", "--limit-rate=10m", "ftp://" + s + "/" + Path, "-O", Path} //filepath.Base(Path)}
	fireParams := []string{"--host", d, "-t", "2s"} //, "--dir", filepath.Dir(Path)}
	params := append(fireParams, wgetParams...)
	//fmt.Println(params)
    cmd := exec.Command("fire", append(fireParams, "rm", "-f", Path)...)
    err := cmd.Run()
    if err != nil {
       goto OK_JUDGE
    }
	cmd = exec.Command("fire", params...)
	_, err = cmd.CombinedOutput()

OK_JUDGE:
	var ok = (err == nil)
	if ok {
        fmt.Println(d, "SUCC")
		//beelog.Info("Succ copy from", s, "to", d)
		left -= 1 // TODO: maybe need lock
		push(src, s)
		push(src, d)
	} else {
        fmt.Println(d, "FAIL")
		//beelog.Warn("Fail copy from", s, "to", d)
		push(src, s)
		left -= 1
		//push(dst, d)
	}
}

// start do distribution
func start() {
	beelog.Info("start to copy")
	go initSource(Source)
	go initDest(Dest)
	var s, d string
	for left != 0 {
		select {
		case d = <-dst:
			s = <-src
			go copywork(s, d)
		default:
			runtime.Gosched()
		}
	}
	log.Println("FINISH")
}

func main() {
	beelog.SetLevel(beelog.LevelInfo)
	var opts struct {
		Source   []string `short:"s" long:"src" description:"source host"`
		Dest     []string `short:"d" long:"dst" description:"destination host"`
		DestFile string   `long:"df" description:"destination file"`
		Path     string   `short:"p" long:"path" description:"file path" default:"/home/work/a"`
	}
	_, err := flags.Parse(&opts)
	if err != nil {
		return
	}
	if opts.DestFile != "" {
		data, err := ioutil.ReadFile(opts.DestFile)
		if err != nil {
			beelog.Error(err)
			return
		}
		Dest = strings.Fields(string(data))
	}
	Source = append(Source, opts.Source...)
	Dest = append(Dest, opts.Dest...)
	Path = opts.Path

	beelog.Debug("dest   :", opts.Dest)
	beelog.Debug("sources:", Source)
	beelog.Info("path   :", opts.Path)

    var confirm string
    fmt.Print("confirm y/n:? ")
    fmt.Fscanf(os.Stdin, "%s", &confirm)
    if strings.TrimSpace(confirm) != "y" {
        os.Exit(0)
    }

	startTime := time.Now()
	start()
	log.Println("Time spend", time.Since(startTime).String())
}
