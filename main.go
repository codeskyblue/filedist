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
	"path/filepath"
	"runtime"
	"strings"
	"time"
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
	log.Println("Total: ", left)
	for _, d := range D {
		beelog.Trace("add dst", d)
		dst <- d
	}
	beelog.Trace("dst done")
}
func push(ch chan string, data string) {
	go func() {
		ch <- data
	}()
}

// file copy function
func copywork(s string, d string) {
	//time.Sleep(time.Second * 1)
	//fmt.Println("jetfire", "-host", d, "-u", "work", "-dir", filepath.Dir(Path),
	beelog.Info("copywork", s, "--->", d)
	cmd := exec.Command("jetfire", "-host", d, "-u", "work", "-dir", filepath.Dir(Path),
		"wget", "-nv", "--limit-rate=10m", "ftp://"+s+"/"+Path, "-O", filepath.Base(Path))
	out, err := cmd.CombinedOutput()
	if false {
		fmt.Print(string(out))
	}
	var ok = (err == nil)
	if ok {
		beelog.Debug("Succ copy from", s, "to", d)
		left -= 1 // TODO: maybe need lock
		push(src, s)
		push(src, d)
	} else {
		beelog.Warn("Fail copy from", s, "to", d)
		push(src, s)
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
	beelog.Debug("path   :", opts.Path)

	startTime := time.Now()
	start()
	log.Println("Time spend", time.Since(startTime).String())
}
