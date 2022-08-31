package cmd

import "github.com/jessevdk/go-flags"

type options struct {
	SStablePath  string `short:"s" long:"sstable" description:"A path to sstable directory" required:"true"`
	Vlog         string `short:"v"  description:"A path to vlog file" required:"true"`
	Checkpoint   string `short:"c" long:"checkpoint"  description:"A path to checkpoint file" required:"true"`
	MemtableSize int    `short:"m" long:"memtable" description:"size of memtable" default:"20"`
}

func Parse() (*options, error) {
	options := options{}
	_, err := flags.Parse(&options)
	if err != nil {
		return nil, err
	}
	return &options, nil

}
