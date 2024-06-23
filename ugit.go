/*
@author: sk
@date: 2024/6/18
*/
package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

func Init() {
	// 初始化文件夹
	Mkdir(GitDir)
	Mkdir(filepath.Join(GitDir, Objects))
	Mkdir(filepath.Join(GitDir, Refs))
	Mkdir(filepath.Join(GitDir, Refs, Tags))
	Mkdir(filepath.Join(GitDir, Refs, Heads))
	// 初始化分支
	oid := Commit("git init", "")
	Branch("master", oid)
	Checkout("master")
}

func HashObject(data []byte, type0 string) string {
	h := sha1.New()
	h.Write(data)
	oid := hex.EncodeToString(h.Sum(nil))
	MustTrue(len(type0) == 4, "err type %s", type0)
	data = append([]byte(type0), data...)
	WriteFile(filepath.Join(GitDir, Objects, oid), data)
	return oid
}

func GetObject(oid string, type0 string) []byte { // cat_file 实现依赖
	data := ReadFile(filepath.Join(GitDir, Objects, oid))
	if len(type0) > 0 {
		MustTrue(type0 == string(data[:TypeLen]), "type miss match %s != %s", type0, string(data[:TypeLen]))
	}
	return data[TypeLen:]
}

func WriteTree(path string) string {
	if len(path) == 0 {
		path = "./"
	}
	items := ReadDir(path)
	buff := bytes.Buffer{}
	for _, item := range items {
		path0 := filepath.Join(path, item.Name())
		if NeedIgnore(path0) {
			continue
		}
		if item.IsDir() {
			buff.WriteString(TypeTree)
			buff.WriteString(WriteTree(path0))
		} else {
			buff.WriteString(TypeBlob)
			data := ReadFile(path0)
			buff.WriteString(HashObject(data, TypeBlob))
		}
		buff.WriteString(item.Name())
		buff.WriteRune('\n')
	}
	return HashObject(buff.Bytes(), TypeTree)
}

func NeedIgnore(path string) bool {
	list := strings.Split(path, "/")
	for _, item := range list {
		if item[0] == '.' {
			return true
		}
	}
	return false
}

func ReadTree(oid string) {
	RemoveAllFile("")
	res := GetTree(oid, "")
	for oid0, path := range res {
		data := GetObject(oid0, TypeBlob)
		WriteFile(path, data)
	}
}

func RemoveAllFile(path string) {
	if len(path) == 0 {
		path = "./"
	}
	items := ReadDir(path)
	for _, item := range items {
		path0 := filepath.Join(path, item.Name())
		if NeedIgnore(path0) {
			continue
		}
		RemoveFile(path0)
	}
}

func GetTree(oid string, path string) map[string]string {
	res := make(map[string]string) // path -> oid
	data := GetObject(oid, TypeTree)
	items := bytes.Split(data, []byte("\n"))
	for _, item := range items {
		if len(item) == 0 {
			continue
		}
		type0 := string(item[:TypeLen])
		oid0 := string(item[TypeLen : TypeLen+OIDLen])
		name := string(item[TypeLen+OIDLen:])
		switch type0 {
		case TypeTree:
			temp := GetTree(oid0, filepath.Join(path, name))
			for key, val := range temp {
				res[key] = val
			}
		case TypeBlob:
			res[oid0] = filepath.Join(path, name)
		default:
			panic(fmt.Sprintf("unknown type: %s", type0))
		}
	}
	return res
}

func Commit(msg string, other string) string { // other用于合并分支的提交一般为空
	buff := bytes.Buffer{}
	buff.WriteString(HdrTree)
	buff.WriteString(WriteTree(""))
	buff.WriteRune('\n')
	buff.WriteString(HdrTime)
	buff.WriteString(time.Now().Format(TimeFormat))
	buff.WriteRune('\n')
	buff.WriteString(HdrAuthor)
	buff.WriteString(GetUser())
	buff.WriteRune('\n')
	head := GetHead()
	if len(head) > 0 {
		buff.WriteString(HdrParent)
		buff.WriteString(head)
		buff.WriteRune('\n')
	}
	if len(other) > 0 {
		buff.WriteString(HdrParent)
		buff.WriteString(other)
		buff.WriteRune('\n')
	}

	buff.WriteRune('\n')
	buff.WriteString(msg)
	oid := HashObject(buff.Bytes(), TypeCommit)
	SetHeadOID(oid)
	return oid
}

func SetHeadOID(oid string) {
	path := TransPath(filepath.Join(GitDir, Head))
	if !strings.HasPrefix(path, GitDir) { // 可能直接引用
		path = filepath.Join(GitDir, Refs, Heads, path)
	}
	WriteFile(path, []byte(oid)) // 这里必定指定 branch
}

func SetHeadRef(ref string) {
	path := filepath.Join(GitDir, Head)
	WriteFile(path, []byte(RefPrefix+ref))
}

func GetHead() string {
	path := filepath.Join(GitDir, Head)
	if FileExist(path) {
		data := ReadFile(path)
		return TransRef(string(data))
	}
	return ""
}

func TransRef(raw string) string {
	if strings.HasPrefix(raw, RefPrefix) {
		return GetRef(raw[len(RefPrefix):]) // 内部会再次调用这个形成递归
	}
	return raw
}

func Log(oid string) {
	if len(oid) == 0 {
		oid = GetHead()
	}
	for {
		commit := GetCommit(oid)
		fmt.Printf("%s\t%s\t%s\t%s\n", oid, commit.Time.Format(TimeFormat), commit.Author, commit.Msg)
		if len(commit.Parent) > 0 {
			oid = commit.Parent[0] // 暂时先不管 合并来的 other分支，以主分支为准 看来源
		} else {
			break
		}
	}
}

func GetCommit(oid string) *Commit0 {
	data := GetObject(oid, TypeCommit)
	items := bytes.Split(data, []byte("\n"))
	i := 0
	res := &Commit0{}
	for len(items[i]) > 0 {
		type0 := string(items[i][:TypeLen])
		switch type0 {
		case HdrTree:
			res.Tree = string(items[i][HdrLen:])
		case HdrTime:
			res.Time = TimeParse(string(items[i][HdrLen:]))
		case HdrAuthor:
			res.Author = string(items[i][HdrLen:])
		case HdrParent:
			res.Parent = append(res.Parent, string(items[i][HdrLen:]))
		default:
			panic(fmt.Sprintf("unknown type: %s", type0))
		}
		i++
	}
	msg := bytes.Buffer{}
	for i < len(items) {
		msg.Write(items[i])
		i++
	}
	res.Msg = strings.TrimSpace(msg.String())
	return res
}

func Checkout(name string) { // 可以切换到 branch 或 tag
	oid := ParseOid(name)
	commit := GetCommit(oid)
	ReadTree(commit.Tree)
	if IsBranch(name) { // 若是分支 HEAD 指向分支  移动时带着分支一起移动
		SetHeadRef(name)
	} else {
		SetHeadOID(oid) // tag HEAD 指向 oid tag 不会移动，对于 tag最好不要切过来修改，只查看，当只读存档使用
	}
}

func IsBranch(name string) bool {
	if strings.Contains(name, Heads) {
		return true
	}
	return FileExist(filepath.Join(GitDir, Refs, Heads, name))
}

func Tag(name, oid string) {
	if len(oid) == 0 {
		oid = GetHead()
	}
	SetRefOID(filepath.Join(Tags, name), oid)
}

func Branch(name, oid string) {
	if len(name) == 0 {
		heads := IterHeads()
		curr := TransPath(filepath.Join(GitDir, Head))
		for _, head := range heads {
			if curr == head {
				fmt.Printf("* %s\n", head)
			} else {
				fmt.Printf("  %s\n", head)
			}
		}
		return
	}
	if len(oid) == 0 {
		oid = GetHead()
	}
	SetRefOID(filepath.Join(Heads, name), oid)
}

func IterHeads() []string {
	res := make([]string, 0)
	err := filepath.Walk(filepath.Join(GitDir, Refs, Heads), func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return err
		}
		index := strings.Index(path, Heads)
		res = append(res, path[index+6:])
		return err
	})
	HandleErr(err)
	return res
}

func SetRefOID(ref, oid string) { // 更新其 oid
	path := filepath.Join(GitDir, Refs, ref)
	path = TransPath(path)
	if !strings.HasPrefix(path, filepath.Join(GitDir, Refs)) { // 转换后的路径不是全路径
		path = filepath.Join(GitDir, Refs, Heads, path)
	}
	dir := filepath.Dir(path)
	Mkdir(dir) // 确保目录存在
	WriteFile(path, []byte(oid))
}

func SetRefRef(ref, refV string) { // 更新其引用
	path := filepath.Join(GitDir, Refs, ref)
	dir := filepath.Dir(path)
	Mkdir(dir) // 确保目录存在
	WriteFile(path, []byte(RefPrefix+refV))
}

func TransPath(path string) string { // 找最终引用 oid 的路径，方便后面修改
	if !FileExist(path) { // 不存在新建的
		return path
	}
	raw := ReadFile(path)
	for strings.HasPrefix(string(raw), RefPrefix) {
		path = string(raw[len(RefPrefix):])
		raw = ReadFile(filepath.Join(GitDir, Refs, Heads, path))
	}
	return path
}

func GetRef(ref string) string {
	paths := []string{filepath.Join(GitDir, Refs, ref), filepath.Join(GitDir, Refs, Tags, ref),
		filepath.Join(GitDir, Refs, Heads, ref)}
	for _, path := range paths { // 从 3 处分别获取
		if FileExist(path) {
			data := ReadFile(path)
			return TransRef(string(data))
		}
	}
	return ""
}

func DelRef(ref string) {
	paths := []string{filepath.Join(GitDir, Refs, ref), filepath.Join(GitDir, Refs, Tags, ref),
		filepath.Join(GitDir, Refs, Heads, ref)}
	for _, path := range paths { // 从 3 处分别获取
		if FileExist(path) {
			DelFile(path)
			break
		}
	}
}

func ParseOid(name string) string {
	if name == Head || name == HeadAlias {
		return GetHead()
	}
	oid := GetRef(name)
	if len(oid) > 0 {
		return oid
	}
	return name
}

func K() { // 用于可视化输出
	res := IterRefs()
	for ref, oid := range res {
		paths := GetPath(oid)
		slices.Reverse(paths)
		for i, path := range paths {
			if i > 0 {
				fmt.Print(" -> ")
			}
			fmt.Print(path)
		}
		fmt.Printf("\t%s\n", ref)
	}
}

func GetPath(oid string) []string {
	res := make([]string, 0)
	for len(oid) > 0 {
		commit := GetCommit(oid)
		res = append(res, oid)
		oid = commit.Parent[0] // 以主流程为准 先不管合并来的分支
	}
	return res
}

func IterRefs() map[string]string {
	res := make(map[string]string)
	res[Head] = ParseOid(Head)
	err := filepath.Walk(filepath.Join(GitDir, Refs), func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return err
		}
		index := strings.Index(path, Refs)
		path = path[index+5:]
		res[path] = ParseOid(path)
		return err
	})
	HandleErr(err)
	return res
}

func Status() {
	Show("")
}

func Reset(oid string) {
	SetHeadOID(oid) // 移动 头与 当前分支到某一次提交
}

func Show(oid string) {
	if len(oid) == 0 {
		oid = GetHead()
	}
	// 显示基本信息
	commit := GetCommit(oid)
	fmt.Printf("commit %s\n", oid)
	fmt.Println(commit.Msg)
	// 显示相比上个版本的修改信息
	currTree := commit.Tree
	parentTree := ""
	if len(commit.Parent[0]) > 0 { // 只看关键链路
		commit = GetCommit(commit.Parent[0])
		parentTree = commit.Tree
	}
	res1 := IterTree(currTree, "")
	res2 := IterTree(parentTree, "")
	// 修改
	for path, oid1 := range res1 {
		if oid2, ok := res2[path]; ok {
			if oid1 != oid2 {
				fmt.Printf("file %s change\n", path)
			}
		}
	}
	for path := range res1 {
		if _, ok := res2[path]; !ok {
			fmt.Printf("file %s add\n", path)
		}
	}
	for path := range res2 {
		if _, ok := res1[path]; !ok {
			fmt.Printf("file %s delete\n", path)
		}
	}
}

func IterTree(oid string, path string) map[string]string {
	if len(oid) == 0 {
		return make(map[string]string)
	}
	// 获取 文件->oid 的映射
	res := make(map[string]string)
	data := GetObject(oid, TypeTree)
	items := bytes.Split(data, []byte("\n"))
	for _, item := range items {
		if len(item) < 4 {
			continue
		}
		path0 := filepath.Join(path, string(item[TypeLen+OIDLen:]))
		oid0 := string(item[TypeLen : TypeLen+OIDLen])
		switch string(item[:TypeLen]) {
		case TypeTree:
			temp := IterTree(oid0, path0)
			for key, val := range temp {
				res[key] = val
			}
		case TypeBlob:
			res[path0] = oid0
		}
	}
	return res
}

func Diff(oid, path string) {
	if len(oid) == 0 {
		oid = GetHead()
	}
	// 找对应的提交树
	commit := GetCommit(oid)
	currTree := commit.Tree
	parentTree := ""
	if len(commit.Parent) > 0 {
		commit = GetCommit(commit.Parent[0]) // 只看主分支的 不看 合并来的分支
		parentTree = commit.Tree
	}
	// 获取对应文件的 oid
	res1 := IterTree(currTree, "")
	res2 := IterTree(parentTree, "")
	oid1, ok1 := res1[path]
	oid2, ok2 := res2[path]
	if !ok1 || !ok2 {
		fmt.Printf("file %s not is change can't diff\n", path)
		return
	}
	// 进行文件对比
	blob1 := GetObject(oid1, TypeBlob)
	blob2 := GetObject(oid2, TypeBlob)
	lines := DiffFile(blob1, blob2)
	blocks := MergeLine(lines)
	for _, block := range blocks {
		if block.Type == ChangeAdd {
			fmt.Printf("add lines \n%s\n", string(block.Content))
		} else if block.Type == ChangeDelete {
			fmt.Printf("delete lines \n%s\n", string(block.Content))
		}
	}
}

func MergeLine(lines []*DiffLine) []*DiffLine {
	if len(lines) == 0 {
		return make([]*DiffLine, 0)
	}
	type0 := lines[0].Type
	buff := &bytes.Buffer{}
	buff.Write(lines[0].Content)
	res := make([]*DiffLine, 0)
	for i := 1; i < len(lines); i++ {
		if lines[i].Type == type0 {
			buff.WriteRune('\n')
			buff.Write(lines[i].Content)
		} else {
			res = append(res, &DiffLine{
				Type:    type0,
				Content: buff.Bytes(),
			})
			type0 = lines[i].Type
			buff = &bytes.Buffer{}
			buff.Write(lines[i].Content)
		}
	}
	res = append(res, &DiffLine{
		Type:    type0,
		Content: buff.Bytes(),
	})
	return res
}

type DiffLine struct {
	Content []byte
	Type    string
}

func DiffFile(blob1, blob2 []byte) []*DiffLine {
	items1 := bytes.Split(blob1, []byte("\n"))
	items2 := bytes.Split(blob2, []byte("\n"))
	// 计算 dp
	dps := make([][]int, len(items1)+1)
	for i := 0; i < len(dps); i++ {
		dps[i] = make([]int, len(items2)+1)
	}
	for i := 1; i <= len(items1); i++ {
		for j := 1; j <= len(items2); j++ {
			if string(items1[i-1]) == string(items2[j-1]) {
				dps[i][j] = dps[i-1][j-1] + 1
			} else {
				dps[i][j] = Max(dps[i-1][j], dps[i][j-1])
			}
		}
	}
	// 根据 dp反推 diff
	res := make([]*DiffLine, 0)
	i1, i2 := len(items1), len(items2)
	for i1 > 0 && i2 > 0 {
		// 先判断增删 最后再判断不变
		if dps[i1][i2] == dps[i1-1][i2] {
			res = append(res, &DiffLine{
				Content: items1[i1-1],
				Type:    ChangeAdd,
			})
			i1--
		} else if dps[i1][i2] == dps[i1][i2-1] {
			res = append(res, &DiffLine{
				Content: items2[i2-1],
				Type:    ChangeDelete,
			})
			i2--
		} else {
			res = append(res, &DiffLine{
				Content: items1[i1-1],
				Type:    ChangeNone,
			})
			i1--
			i2--
		}
	}
	for i1 > 0 {
		res = append(res, &DiffLine{
			Content: items1[i1-1],
			Type:    ChangeAdd,
		})
		i1--
	}
	for i2 > 0 {
		res = append(res, &DiffLine{
			Content: items2[i2-1],
			Type:    ChangeDelete,
		})
		i2--
	}
	slices.Reverse(res)
	return res
}

func Merge(branch string) { // 把当前分支合并到目标分支 并指向最终合并内容
	oid := ParseOid(branch)
	commit := GetCommit(oid)
	res := GetTree(commit.Tree, "")
	for oid0, path := range res {
		data := GetObject(oid0, TypeBlob)
		if FileExist(path) { // 这里采用两路合并 取多不取少
			src := ReadFile(path)
			temp := MergeFile(src, data) // 都存在进行 merge
			WriteFile(path, temp)
		} else {
			WriteFile(path, data) // 文件不存在直接添加，对应分支删除的文件先不管
		}
	}
	Commit("merge "+branch, oid)
}

func MergeFile(src []byte, dst []byte) []byte {
	lines := DiffFile(dst, src)
	hasDel := false
	for _, line := range lines {
		if line.Type == ChangeDelete {
			hasDel = true
			break
		}
	}
	if !hasDel { // 如果目标代码没有删除行为可以直接合并，不用用户解冲突了
		return dst
	}
	blocks := MergeLine(lines)
	buff := &bytes.Buffer{}
	for _, block := range blocks {
		switch block.Type {
		case ChangeNone:
			buff.Write(block.Content)
		case ChangeAdd:
			buff.WriteString("<<<<<<< ADD\n")
			buff.Write(block.Content)
			buff.WriteString("\n>>>>>>>")
		case ChangeDelete:
			buff.WriteString("<<<<<<< DELETE\n")
			buff.Write(block.Content)
			buff.WriteString("\n>>>>>>>")
		}
		buff.WriteRune('\n')
	}
	return buff.Bytes()
}

// 暂时不支持按分支拉取  实际实现应该允许按分支拉取，且对于 objects 只拉取需要的
func Fetch() { // 这里暂时不接入远程存储，先以本地存储为主  在  /Users/${用户家目录}/.ugit/${project} 下存储远程信息
	dir := RemoteDir()
	// 文件夹不存在进行初始化
	//if !FileExist(dir) {
	//
	//}
	Mkdir(dir)
	Mkdir(filepath.Join(dir, Remotes))
	Mkdir(filepath.Join(dir, Objects))
	Mkdir(filepath.Join(GitDir, Refs, Remotes)) // 本地也要初始化
	// 处理分支 分支必须全部拉取为最新的
	path := filepath.Join(dir, Remotes)
	err := filepath.Walk(path, func(sub string, info fs.FileInfo, err error) error {
		if sub == path { // 当前文件夹先忽略吧，一定是有的
			return err
		}
		index := strings.LastIndex(sub, Remotes)
		sub = sub[index+8:]
		if info.IsDir() {
			Mkdir(filepath.Join(GitDir, Refs, Remotes, sub))
		} else {
			data := ReadFile(filepath.Join(path, sub))
			WriteFile(filepath.Join(GitDir, Refs, Remotes, sub), data)
		}
		return err
	})
	HandleErr(err)
	// 处理文件 文件有 oid 标识，可以优化，已经存在的不要再拉取了，这里简单起见，全部拉取
	path = filepath.Join(dir, Objects)
	err = filepath.Walk(path, func(path0 string, info fs.FileInfo, err error) error {
		if info.IsDir() { // 不处理文件夹
			return err
		}
		data := ReadFile(path0)
		index := strings.LastIndex(path0, Objects) // 都是文件没有目录
		path0 = filepath.Join(GitDir, path0[index:])
		WriteFile(path0, data)
		return err
	})
	HandleErr(err)
}

func Push(branch string) { // 按分支推送
	if len(branch) == 0 { // 默认就是当前分支
		branch = TransPath(filepath.Join(GitDir, Head))
	}
	dir := RemoteDir()
	// 文件夹不存在，进行初始化
	//if !FileExist(dir) {
	//
	//}
	Mkdir(dir)
	Mkdir(filepath.Join(dir, Remotes))
	Mkdir(filepath.Join(dir, Objects))
	Mkdir(filepath.Join(GitDir, Refs, Remotes)) // 本地也要初始化
	oid := ParseOid(branch)
	path := filepath.Join(dir, Remotes, branch)
	if FileExist(path) { // 对应分支不存在可以直接提交，否则远程对应分支oid必须是当前分支 oid的父节点
		remoteOid := string(ReadFile(path)) // 主要是防止自己冲掉别人的数据
		if !IsParent(oid, remoteOid) {
			fmt.Printf("please merge remotes/%s first", branch)
			return
		}
	}
	// 处理分支
	data := ReadFile(filepath.Join(GitDir, Refs, Heads, branch))
	WriteFile(filepath.Join(GitDir, Refs, Remotes, branch), data)
	WriteFile(filepath.Join(dir, Remotes, branch), data)
	// 数据处理  正常来说应该按提交树把数据上传的，这里全部上传
	path = filepath.Join(GitDir, Objects)
	err := filepath.Walk(path, func(path0 string, info fs.FileInfo, err error) error {
		if info.IsDir() { // 不处理文件夹
			return err
		}
		data = ReadFile(path0)
		index := strings.LastIndex(path0, Objects) // 都是文件没有目录
		path0 = filepath.Join(dir, path0[index:])
		WriteFile(path0, data)
		return err
	})
	HandleErr(err)
}

func IsParent(son string, parent string) bool {
	sons := []string{son}
	for len(sons) > 0 {
		son = sons[0]
		sons = sons[1:]
		if son == parent {
			return true
		}
		commit := GetCommit(son)
		sons = append(sons, commit.Parent...)
	}
	return false
}
