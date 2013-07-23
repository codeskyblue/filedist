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

func WgetTimeout(path string, addr string, timeout time.Duration, params ...string) (md5sum string, err error) {
	fi, err := os.Create(path)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			os.Remove(path)
		}
	}()
	defer fi.Close()

	h := md5.New()
	mw := io.MultiWriter(fi, h)

	wcmd := exec.Command("wget", append(params, "-O-", addr)...)
	wcmd.Stdout = mw
	wcmd.Stderr = os.Stderr
	wcmd.Timeout = timeout
	err = wcmd.Run()
	if err != nil {
		return
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
