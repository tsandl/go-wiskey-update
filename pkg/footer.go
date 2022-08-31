package wiskey

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
)

const (
	footerSize = 4 //how many bytes are in the footer(indexOffset)
)

//footer in the sstable file, it shows where the index starts in the file
type Footer struct {
	indexOffset uint32 // the Offset where indexes starts
}

func DefaultFooter() *Footer {
	return &Footer{
		indexOffset: 0,
	}
}

func NewFooter(buffer []byte) *Footer {
	if len(buffer) != footerSize {
		panic("Invalid header length")
	}
	offset := binary.BigEndian.Uint32(buffer[:footerSize])
	return &Footer{indexOffset: offset}
}

//save the header in the given writeCloser
func (h *Footer) writeTo(writer io.Writer) int {
	offset, err := writer.Write(h.asByteArray())
	if err != nil {
		panic(err)
	}
	return offset
}

//convert header to binary array
func (h *Footer) asByteArray() []byte {
	buffer := bytes.NewBuffer(make([]byte, 0, footerSize))
	err := binary.Write(buffer, binary.BigEndian, h)
	if err != nil {
		panic(err)
	}
	return buffer.Bytes()
}


//Read the footer
func readFooter(stats os.FileInfo, reader *os.File) *Footer {
	buf := make([]byte, footerSize)
	reader.Seek(stats.Size()-footerSize, 0)
	reader.Read(buf)
	return NewFooter(buf)
}
