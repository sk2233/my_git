/*
@author: sk
@date: 2024/6/17
*/
package main

const (
	GitDir    = ".ugit"
	Objects   = "objects"
	Head      = "HEAD"
	Refs      = "refs"
	Tags      = "tags"
	Heads     = "heads"
	HeadAlias = "@"
	Remotes   = "remotes"
)

const (
	// 固定长度为 4
	TypeBlob   = "blob"
	TypeTree   = "tree"
	TypeCommit = "cmit"
)

const (
	TypeLen = 4
	OIDLen  = 40
	HdrLen  = 4
)

const ( // commit 各种头信息标识
	HdrTree   = "tree"
	HdrTime   = "time"
	HdrAuthor = "ator"
	HdrParent = "part"
)

const (
	TimeFormat = "2006-01-02 15:04:05"
)

const (
	RefPrefix = "ref:"
)

const (
	ChangeNone   = "ChangeNone"
	ChangeAdd    = "ChangeAdd"
	ChangeDelete = "ChangeDelete"
)
