package wiskey

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestVlog_Append(t *testing.T) {
	file, _ := ioutil.TempFile("", "")
	checkpoint, _ := ioutil.TempFile("", "")
	defer file.Close()
	defer checkpoint.Close()
	defer os.Remove(file.Name())
	defer os.Remove(checkpoint.Name())
	vlog := NewVlog(file.Name(), checkpoint.Name())
	//test entries
	entries := FakeEntries()
	//save entries
	for _, entry := range entries {
		length := uint32(uint32Size /*key length*/ + uint32Size /*value length*/ + len(entry.key) /*ANITA takes 5 bytes*/ + len(entry.value) /*DEVELOPER takes 8 bytes*/)
		meta, err := vlog.Append(&entry)
		if err != nil {
			t.Error(err)
		}
		if meta.length != length {
			t.Error("The lengths don't match")
		}
	}
	currentOffset := uint32(0)
	//search them
	for _, entry := range entries {
		length := uint32(uint32Size /*key length*/ + uint32Size /*value length*/ + len(entry.key) /*ANITA takes 5 bytes*/ + len(entry.value) /*DEVELOPER takes 8 bytes*/)
		val, err := vlog.Get(ValueMeta{length: length, offset: currentOffset})
		if err != nil {
			t.Error(err)
		}
		currentOffset += length
		if string(val.key) != string(entry.key) {
			t.Error("Wrong keys")
		}
		if string(val.value) != string(entry.value) {
			t.Error("Wrong values")
		}
	}
}
