/*
@author: sk
@date: 2024/6/18
*/
package main

import "os"

// https://www.leshenko.net/p/ugit/#

func main() {
	cmd := os.Args[1]
	HandleCmd(cmd, os.Args[2:])
	//Merge("remotes/master")
	//Commit("change main")
	//Init()
	//Log("")
	//Tag("feat/cool", "")
	//K()
	//Branch("feat/old", "")
	//Status()
	//Branch("", "")
	//Show("")
	//Diff("", "main.go")
	//Status()
	//curr := ReadFile("test/file2")
	//old := ReadFile("test/file1")
	//res := DiffFile(curr, old)
	//out := &bytes.Buffer{}
	//for _, item := range res {
	//	switch item.Type {
	//	case ChangeNone:
	//		out.Write(item.Content)
	//	case ChangeAdd:
	//		out.Write([]byte("===========ADD==========\n"))
	//		out.Write(item.Content)
	//		out.Write([]byte("\n===========END=========="))
	//	case ChangeDelete:
	//		out.Write([]byte("===========DELETE==========\n"))
	//		out.Write(item.Content)
	//		out.Write([]byte("\n===========END=========="))
	//	}
	//	out.Write([]byte("\n"))
	//}
	//WriteFile("test/out", out.Bytes())
	//fmt.Println(os.UserHomeDir())
	//filepath.Walk(".ugit", func(path string, info fs.FileInfo, err error) error {
	//	return err
	//})
	//Push("")
	//Fetch()
}
