package main

import (
	"fmt"
	"github.com/shxsun/flags"
	"log"
	"strings"
	"time"
)

var opts struct {
	Timeout   string `short:"t" long:"timeout" description:"down timeout" default:"0s"`
	LimitRate string `short:"l" long:"limit-rate" description:"download speed limit per second" default:"10m"`
	Md5sum    string `short:"m" long:"md5sum" description:"check if md5sum matches"`
	Wget      string `long:"wget" description:"specfity which wget to use" default:"/usr/bin/wget"`
}

func main() {
	p := flags.NewParser(&opts, flags.Default^flags.PassArgument)
	p.Usage = "[OPTIONS]  <URL> <target>"
	params := []string{}
	args, err := p.Parse()
	if err != nil {
		return
	}
	if len(args) != 2 {
		log.Fatal("Use --help for more help")
	}
	timeout, err := time.ParseDuration(opts.Timeout)
	if err != nil {
		log.Fatal(err)
	}
	params = append(params, "--limit-rate="+opts.LimitRate)

	md5sum, err := WgetTimeout(args[1], args[0], timeout, params...)
	if err != nil {
		log.Fatal(err)
	}
	checkMd5sum := strings.ToLower(opts.Md5sum)
	if checkMd5sum != "" {
		if checkMd5sum != md5sum {
			fmt.Printf("expect (%s) but got (%s)\n", checkMd5sum, md5sum)
		} else {
			fmt.Println("md5 matches")
		}
	} else {
		fmt.Println(md5sum)
	}
}
