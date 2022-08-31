package wiskey

import (
	"bytes"
	"encoding/binary"
	"os"
)

type indexes []tableIndex

const (
	sstableExtension  = ".sstable$"
	sstableFileLength = 10
)

type SSTable struct {
	footer  *Footer
	indexes indexes
	reader  *os.File
	log     *vlog
}

//Constructor
func ReadTable(reader *os.File, log *vlog) *SSTable {
	stats, _ := reader.Stat()
	//read footer
	footer := readFooter(stats, reader)
	indexes := readIndexes(stats, reader, *footer)
	return &SSTable{footer: footer, indexes: indexes, reader: reader, log: log}
}

func OverrideVlogOffset(position int, meta *ValueMeta, file *os.File) error {
	stats, _ := file.Stat()
	footer := readFooter(stats, file)
	indexes := readIndexes(stats, file, *footer)
	index := indexes[position]
	buffer := bytes.NewBuffer([]byte{})
	//offset
	if err := binary.Write(buffer, binary.BigEndian, meta.offset); err != nil {
		return err
	}
	//length
	if err := binary.Write(buffer, binary.BigEndian, meta.length); err != nil {
		return err
	}
	_, err := file.WriteAt(buffer.Bytes(), int64(index.Offset))
	if err != nil {
		return err
	}
	return nil
}

func (table *SSTable) Close() {
	table.reader.Close()
}

func (table *SSTable) Get(key []byte) (*SearchEntry, bool) {
	//try smallest key
	firstIndex := table.indexes[0]
	compare, value := table.find(key, firstIndex)
	if compare == 0 {
		return value, true
	}
	//if smaller than firstIndex key in file than key is not in the file
	if compare < 0 {
		return nil, false
	}
	//try biggest key
	//TODO: block can contain multiple keys, need to check them all
	lastIndex := table.indexes[len(table.indexes)-1]
	compare, value = table.find(key, lastIndex)
	if compare == 0 {
		return value, true
	}
	//if bigger than lastIndex key in file than key is not in the file
	if compare > 0 {
		return nil, false
	}
	search, found, _ := table.binarySearch(key)
	return search, found
}

func (table *SSTable) KeyAtIndex(key []byte) (bool, int) {
	_, found, index := table.binarySearch(key)
	return found, index
}

//Tries to find given key in the sstable
//Returns 1. value byte array or nil if not found
//2. bool true if found,false otherwise
//3. at which index this key was found
func (table *SSTable) binarySearch(key []byte) (*SearchEntry, bool, int) {
	left := 0
	right := len(table.indexes) - 1
	for left < right {
		middle := (right-left)/2 + left
		index := table.indexes[middle]
		//read key length
		tableReader := NewReader(table.reader, int64(index.Offset))
		fileKeyLength := tableReader.readKeyLength()
		//read actual key from the file
		keyBuffer := tableReader.readKey(fileKeyLength)
		compare := bytes.Compare(key, keyBuffer)
		if compare == 0 {
			return table.fetchFromVlog(tableReader), true, middle
		} else if compare > 0 {
			left = middle + 1
		} else {
			right = middle - 1
		}
	}
	index := table.indexes[left]
	tableReader := NewReader(table.reader, int64(index.Offset))
	for tableReader.offset != index.BlockLength {
		keyLength := tableReader.readKeyLength()
		keyFromFile := tableReader.readKey(keyLength)
		if bytes.Compare(key, keyFromFile) == 0 {
			return table.fetchFromVlog(tableReader), true, left
		}
		tableReader.readTimestamp()
		tableReader.readValueOffset()
		tableReader.readValueLength()
	}
	return nil, false, -1
}

func (table *SSTable) fetchFromVlog(tableReader *SSTableReader) *SearchEntry {
	timestamp := tableReader.readTimestamp()
	offset := tableReader.readValueOffset()
	length := tableReader.readValueLength()
	get, err := table.log.Get(ValueMeta{length: length, offset: offset})
	if err != nil {
		panic(err)
	}
	return &SearchEntry{key: get.key, value: get.value, timestamp: timestamp}
}

func (table *SSTable) find(key []byte, index tableIndex) (int, *SearchEntry) {
	searchKeyLength := uint32(len(key))
	tableReader := NewReader(table.reader, int64(index.Offset))
	fileKeyLength := tableReader.readKeyLength()
	//if keys length are not the same then don't make sense to compare an actual key
	if searchKeyLength != fileKeyLength {
		return len(key) - int(fileKeyLength), nil
	}
	//read actual key from the file
	keyBuffer := tableReader.readKey(searchKeyLength)
	compare := bytes.Compare(key, keyBuffer)
	//they are equal
	if compare == 0 {
		return 0, table.fetchFromVlog(tableReader)
	}
	return compare, nil
}

//Read the index from the file to in memory slice
func readIndexes(stats os.FileInfo, reader *os.File, footer Footer) indexes {
	buffer := make([]byte, stats.Size()-int64(footer.indexOffset)-footerSize)
	reader.ReadAt(buffer, int64(footer.indexOffset))
	start := 0
	end := len(buffer)
	indexes := indexes{}
	for start != end {
		blockLength := binary.BigEndian.Uint32(buffer[start : start+4])
		blockOffset := binary.BigEndian.Uint32(buffer[start+4 : start+8])
		indexes = append(indexes, tableIndex{Offset: blockOffset, BlockLength: blockLength})
		start += 8
	}
	return indexes
}

type SearchEntry struct {
	key       []byte
	value     []byte
	timestamp uint64
}
