# go-wiskey

Golang implementation
of [Wiskey](https://www.usenix.org/system/files/conference/fast16/fast16-papers-lu.pdf)
paper

# Description

Dead simple lsm implementation which stores values in the vlog which decreases
write amplification of lsm tree during merging

# Things that are implemented

1. [X] SSTable
    - [X] Create sstable
    - [X] Read from sstable
2. [X] Memtable(in memory redblack tree that stores the data and flushes it once
   memory is full)
    - [X] Put
    - [X] Delete
    - [X] Get
3. [X] Lsm tree
    - [X] Put
    - [X] Get
    - [X] Delete
4. [X] Http interface
    - [X] Http Get
    - [X] Http Put
    - [X] Http Delete
5. [X] Crash recovery
    - [X] Store the last head position in the separate file
    - [X] Store al values from head to tail into the memtable during recovery
6. [X] Merge sstable files
7. [X] Cli interface
    - [X] specify sstable path
    - [X] specify vlog path
    - [X] specify checkpoint path
    - [X] specify memtable size
8. [ ] Reclaim space
    - [X] Merge sstables
    - [ ] Garbage collect vlog

## Install

In order to install the binary run `go get github.com/strogiyotec/go-wiskey` ,
it will be installed in `$HOME/go/bin/wiskey`

## Usage

In order to start the app run
`wiskey -s ../go-wiskey/sstable -v vlog -c checkpoint -m 20`
where :

1. `-s` - directory with sstables
2. `-v` - path to vlog file(vlog doesn't have to exist)
3. `-c` - path to checkpoint (checkpoint doesn't have to exist)
4. `-m` - memtable size in bytes(the size of in memory red black tree that keeps
   keys , when full will flush this tree to sstable)

It will start an http server

### Http server

In order to GET/UPDATE/DELETE you can use http endpoints

1. Save key value
   - `curl -X POST -H "Content-Type: application/json" -d '{"value":"Developer"}' http://localhost:8080/anita`
   it will save value `Developer` with a key `anita`
2. Get by key - `curl -i localhost:8080/fetch/anita`
3. Delete by key - `curl -i localhost:8080/fetch/anita`

### How it works

Here is the general image on how the storage works
![storage](https://raw.githubusercontent.com/strogiyotec/go-wiskey/master/images/Architecture.jpg)
