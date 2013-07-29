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
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

var srcQueue = make(chan string)
var dstQueue = make(chan string)
var Source = []string{}
var Dest = []string{}
var Path string
var left = -1

func initSource(S []string) {
	for _, d := range S {
		beelog.Trace("add source:", d)
		srcQueue <- d
	}
	beelog.Info("src init done")
}
func initDest(D []string) {
	left = len(D)
	log.Println("Total target: ", len(D))
	for _, d := range D {
		beelog.Trace("add dst", d)
		dstQueue <- d
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
	beelog.Debug("copy work", s, d)
	//wgetParams := []string{"wget", "-nv", "--limit-rate=10m", fmt.Sprintf("ftp://%s/%s", s, Path), "-O", Path} //filepath.Base(Path)}
	wgetParams := []string{"wget", "-nv", "--limit-rate=10m", fmt.Sprintf("http://%s:4456/%s", s, Path), "-O", Path} //filepath.Base(Path)}
	fireParams := []string{"--host", d, "-t", opts.Timeout}
	params := append(fireParams, wgetParams...)
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
		push(srcQueue, s)
		push(srcQueue, d)
	} else {
		fmt.Println(d, "FAIL")
		//beelog.Warn("Fail copy from", s, "to", d)
		push(srcQueue, s)
		left -= 1
		//push(dstQueue, d)
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
		case d = <-dstQueue:
			s = <-srcQueue
			go copywork(s, d)
		default:
			runtime.Gosched()
		}
	}
	log.Println("FINISH")
}

var opts struct {
	Source   []string `short:"s" long:"src" description:"source host"`
	Dest     []string `short:"d" long:"dst" description:"destination host"`
	DestFile string   `short:"D" long:"dfile" description:"destination host from file"`
	Path     string   `short:"p" long:"path" description:"file path" default:"/home/work/a"`
	Timeout  string   `short:"t" long:"timeout" description:"for each machine download timeout" default:"10m"`
}

func main() {
	beelog.SetLevel(beelog.LevelInfo)
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
	beelog.Debug(opts)
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
