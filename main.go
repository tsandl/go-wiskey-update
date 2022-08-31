package main

import (
	"go-wiskey/cmd"
	"go-wiskey/http"
	. "go-wiskey/pkg"
)

func main() {
	parse, err := cmd.Parse()
	if err != nil {
		panic(err)
	}
	vlog := NewVlog(parse.Vlog, parse.Checkpoint)
	memtable := NewMemTable(parse.MemtableSize)
	tree := NewLsmTree(vlog, parse.SStablePath, memtable, 120)
	http.Start(tree)
}
