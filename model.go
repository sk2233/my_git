/*
@author: sk
@date: 2024/6/18
*/
package main

import "time"

type Commit0 struct {
	Tree   string
	Time   time.Time
	Author string
	Parent []string
	Msg    string
}
