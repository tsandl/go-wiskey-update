package wiskey

import (
	binary "encoding/binary"
	"io"
	"os"
)

type vlog struct {
	file       string
	size       uint32 // current size of the file,it has to be updated every time you append a new value
	checkpoint string //path to the file with checkpoint
}

func NewVlog(file string, checkpoint string) *vlog {
	vlogFile, err := os.OpenFile(file, os.O_CREATE, 0666)
	vlogFile.Close()
	if err != nil {
		panic(err)
	}
	stat, err := os.Stat(file)
	if err != nil {
		panic(err)
	}
	return &vlog{
		file:       file,
		checkpoint: checkpoint,
		size:       uint32(stat.Size()),
	}
}

//Save the latest vlog head position in the checkpoint file
func (log *vlog) FlushHead() error {
	writer, err := os.OpenFile(log.checkpoint, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer writer.Close()
	err = writer.Truncate(0)
	if err != nil {
		return err
	}
	return binary.Write(writer, binary.BigEndian, log.size)
}

// Example of vlog entry to read
//+------------+--------------+-----+-------+
//| Key Length | Value length | Key | Value |
//+------------+--------------+-----+-------+
func (log *vlog) Get(meta ValueMeta) (*TableEntry, error) {
	reader, err := os.OpenFile(log.file, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	reader.Seek(int64(meta.offset), 0)
	buffer := make([]byte, meta.length)
	reader.Read(buffer)
	keyLength := binary.BigEndian.Uint32(buffer[0:4])
	key := buffer[8 : 8+keyLength]
	value := buffer[8+keyLength:]
	return &TableEntry{key: key, value: value}, nil
}

func (log *vlog) RunGc(entries int, lsm *LsmTree) error {
	file, err := os.OpenFile(log.file, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return err
	}
	logFileSize := stat.Size()
	if logFileSize == 0 {
		return nil
	}
	readBytesSize := int64(0) //how many bytes were read from a file
	counter := 0
	for readBytesSize < logFileSize && counter < entries {
		keyLengthBuffer := make([]byte, uint32Size)
		//read key length
		_, _ = file.Read(keyLengthBuffer)
		keyLength := binary.BigEndian.Uint32(keyLengthBuffer)
		//read value length
		valueLengthBuffer := make([]byte, uint32Size)
		_, _ = file.Read(valueLengthBuffer)
		valueLength := binary.BigEndian.Uint32(valueLengthBuffer)
		keyBuffer := make([]byte, keyLength)
		valueBuffer := make([]byte, valueLength)
		_, _ = file.Read(keyBuffer)
		_, _ = file.Read(valueBuffer)
		tableWithIndexes := lsm.Exists(keyBuffer)
		if len(tableWithIndexes) != 0 {
			entry := &TableEntry{key: keyBuffer, value: valueBuffer}
			valueMeta, err := log.Append(entry)
			if err != nil {
				return err
			}
			for i := range tableWithIndexes {
				tableWithIndex := tableWithIndexes[i]
				file, err := os.OpenFile(tableWithIndex.tablePath, os.O_RDWR, 0666)
				if err != nil {
					return err
				}
				err = OverrideVlogOffset(tableWithIndex.index, valueMeta, file)
				if err != nil {
					return err
				}
				err = file.Close()
				if err != nil {
					return err
				}
			}
		}
		readBytesSize += int64(uint32Size + keyLength + uint32Size + valueLength)
		counter++
	}
	//TODO: so we skipped deleted entries
	//now we have to remove the beginning of the file
	//starting from readBytesSize position
	err = truncateVlog(readBytesSize, log.file)
	if err != nil {
		return err
	}
	info, err := os.Stat(log.file)
	if err != nil {
		return err
	}
	log.size = uint32(info.Size())
	return nil
}

func truncateVlog(offset int64, file string) error {
	fin, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fin.Close()

	tempFileName := RandStringBytes(10)
	fout, err := os.Create(tempFileName)
	if err != nil {
		return err
	}
	defer fout.Close()

	// Offset is the number of bytes you want to exclude
	_, err = fin.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}

	_, err = io.Copy(fout, fin)
	if err != nil {
		return err
	}
	if err := os.Remove(file); err != nil {
		panic(err)
	}

	if err := os.Rename(tempFileName, file); err != nil {
		return err
	}
	return nil
}

//Restore vlog to given memtable
func (log *vlog) RestoreTo(headOffset uint32, memtable *Memtable) error {
	reader, err := os.OpenFile(log.file, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer reader.Close()
	_, err = reader.Seek(int64(headOffset), 0)
	if err != nil {
		return err
	}
	stat, err := reader.Stat()
	if err != nil {
		return err
	}
	length := stat.Size() - int64(headOffset)
	if length == 0 {
		//head is the tail
		return nil
	}
	buffer := make([]byte, length)
	_, err = reader.Read(buffer)
	if err != nil {
		return err
	}
	lastPosition := 0
	nextOffset := uint32(0)
	for lastPosition != len(buffer) {
		keyLength := binary.BigEndian.Uint32(buffer[lastPosition : lastPosition+4])
		valueLength := binary.BigEndian.Uint32(buffer[lastPosition+4 : lastPosition+8])
		key := buffer[lastPosition+8 : lastPosition+8+int(keyLength)]
		metaLength := uint32Size + uint32Size + int(keyLength) + int(valueLength)
		err := memtable.Put(key, &ValueMeta{length: uint32(metaLength), offset: nextOffset + headOffset})
		if err != nil {
			return err
		}
		nextOffset += uint32(metaLength)
		lastPosition += uint32Size
		lastPosition += uint32Size
		lastPosition += int(keyLength)
		lastPosition += int(valueLength)
	}
	log.size = uint32(stat.Size())
	return nil
}

//Append new entry to the head of vlog
//the binary format for entry is [klength,vlength,key,value]
//we store key in vlog for garbage collection purposes
// Example of signle entry in vlog
//+------------+--------------+-----+-------+
//| Key Length | Value length | Key | Value |
//+------------+--------------+-----+-------+
func (log *vlog) Append(entry *TableEntry) (*ValueMeta, error) {
	writer, err := os.OpenFile(log.file, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	defer writer.Close()
	length, err := entry.writeTo(writer)
	if err != nil {
		return nil, err
	}
	meta := &ValueMeta{length: length, offset: log.size}
	log.size += length
	return meta, nil
}

//metadata of saved entry in vlog
type ValueMeta struct {
	length uint32 //value length in vlog file
	offset uint32 //value offset in vlog file
}
