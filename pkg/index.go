package wiskey

import (
	"bytes"
	"encoding/binary"
	"io"
)

//indexes to find an entry in a file
type tableIndex struct {
	Offset      uint32 //Offset of the file where index starts
	BlockLength uint32 //the length of the index
}

//Write index to the end of sstable
//+-------------+--------+
//| BlockLength | Offset |
//+-------------+--------+
func (index *tableIndex) WriteTo(w io.Writer) error {
	buf := bytes.NewBuffer([]byte{})

	if err := binary.Write(buf, binary.BigEndian, index.BlockLength); err != nil {
		return err
	}

	if err := binary.Write(buf, binary.BigEndian, index.Offset); err != nil {
		return err
	}

	_, err := w.Write(buf.Bytes())
	//can be null
	return err
}
