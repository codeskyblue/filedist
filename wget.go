package main

import (
	"crypto/md5"
	"fmt"
	"github.com/shxsun/exec"
	"io"
	"os"
	"time"
)

func Wget(path string, addr string) (md5sum string, err error) {
	return WgetTimeout(path, addr, time.Duration(0))
}

func WgetTimeout(path string, addr string, timeout time.Duration) (md5sum string, err error) {
	defer func() {
		if err != nil {
			fmt.Println(err)
		}
	}()
	fi, err := os.Create(path)
	if err != nil {
		return
	}
	defer fi.Close()

	h := md5.New()
	mw := io.MultiWriter(fi, h)

	wcmd := exec.Command("wget", "-q", "-O-", addr)
	wcmd.Stdout = mw
	wcmd.Timeout = timeout
	err = wcmd.Run()
	if err != nil {
		return
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
