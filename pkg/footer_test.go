package wiskey

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestFooterWriteAndRead(t *testing.T) {
	file, _ := ioutil.TempFile("", "")
	defer file.Close()
	defer os.Remove(file.Name())
	header := Footer{ indexOffset: 100}
	writeToFile(file.Name(), header)
	buf := readFromFile(file.Name())
	headerFromFile := NewFooter(buf)

	if headerFromFile.indexOffset != header.indexOffset {
		t.Error("Index offsets don't match")
	}
}

func readFromFile(fileName string) []byte {
	reader, _ := os.Open(fileName)
	buf := make([]byte, footerSize)
	stats, _ := reader.Stat()
	reader.Seek(stats.Size()-footerSize,0)
	reader.Read(buf)
	reader.Close()
	return buf
}

func writeToFile(fileName string, header Footer) {
	writer, _ := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0600)
	header.writeTo(writer)
	writer.Close()
}
