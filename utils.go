/*
@author: sk
@date: 2024/6/17
*/
package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

func HandleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func Mkdir(path string) {
	err := os.MkdirAll(path, os.ModePerm)
	HandleErr(err)
}

func WriteFile(path string, data []byte) {
	err := os.WriteFile(path, data, 0666)
	HandleErr(err)
}

func DelFile(path string) {
	err := os.Remove(path)
	HandleErr(err)
}

func ReadFile(path string) []byte {
	bs, err := os.ReadFile(path)
	HandleErr(err)
	return bs
}

func MustTrue(val bool, msg string, args ...any) {
	if !val {
		panic(fmt.Sprintf(msg, args...))
	}
}

func ReadDir(path string) []os.DirEntry {
	items, err := os.ReadDir(path)
	HandleErr(err)
	return items
}

func RemoveFile(path string) {
	err := os.RemoveAll(path)
	HandleErr(err)
}

func GetUser() string {
	curr, err := user.Current()
	HandleErr(err)
	return curr.Name
}

func FileExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	panic(err)
}

func TimeParse(time0 string) time.Time {
	res, err := time.Parse(TimeFormat, time0)
	HandleErr(err)
	return res
}

func Max[T int](val1, val2 T) T {
	if val1 > val2 {
		return val1
	}
	return val2
}

func RemoteDir() string {
	// 这里暂时不接入远程存储，先以本地存储为主  在  /Users/${用户家目录}/.ugit/${project} 下存储远程信息
	// 远程目录下主要有  objects remote 两个目录用来存储数据
	dir, err := os.UserHomeDir()
	HandleErr(err)
	wd, err := os.Getwd()
	HandleErr(err)
	index := strings.LastIndex(wd, "/")
	return filepath.Join(dir, GitDir, wd[index+1:])
}
