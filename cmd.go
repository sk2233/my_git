/*
@author: sk
@date: 2024/6/18
*/
package main

import "fmt"

func HandleCmd(cmd string, args []string) {
	switch cmd {
	case "init":
		Init()
	case "commit":
		if len(args) == 0 {
			fmt.Println("commit msg is must")
			return
		}
		Commit(args[0], "")
	case "log":
		oid := ""
		if len(args) > 0 {
			oid = args[0]
		}
		Log(oid)
	case "checkout":
		if len(args) == 0 {
			fmt.Println("checkout need a tag or branch")
			return
		}
		Checkout(args[0])
	case "tag":
		if len(args) == 0 {
			fmt.Println("tag name is must")
			return
		}
		oid := ""
		if len(args) > 1 {
			oid = args[1]
		}
		Tag(args[0], oid)
	case "branch":
		if len(args) == 0 {
			Branch("", "")
			return
		}
		oid := ""
		if len(args) > 1 {
			oid = args[1]
		}
		Branch(args[0], oid)
	case "k":
		K()
	case "status":
		Status()
	case "reset":
		if len(args) == 0 {
			fmt.Println("reset oid is must")
			return
		}
		Reset(args[0])
	case "diff":
		if len(args) == 0 {
			fmt.Println("diff file is must")
			return
		}
		oid := ""
		file := args[0]
		if len(args) > 1 {
			oid = args[0]
			file = args[1]
		}
		Diff(oid, file)
	case "merge":
		if len(args) == 0 {
			fmt.Println("merge branch is must")
			return
		}
		Merge(args[0])
	case "fetch":
		Fetch()
	case "push":
		branch := ""
		if len(args) > 0 {
			branch = args[0]
		}
		Push(branch)
	default:
		fmt.Printf("unknown cmd %s", cmd)
	}
}
