package main

import (
	"github.com/tsandl/go-wiskey-update/cmd"
	"github.com/tsandl/go-wiskey-update/http"
	"github.com/tsandl/go-wiskey-update/pkg"
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
