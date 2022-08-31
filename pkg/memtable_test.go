package wiskey

import (
	"math/rand"
	"testing"
)

const (
	memTableSize = 100 //just a random memtable size
)

func TestMemtable_Put(t *testing.T) {
	table := NewMemTable(memTableSize)
	key := []byte("myKey")
	value := &ValueMeta{length: rand.Uint32(), offset: rand.Uint32()}
	err := table.Put(key, value)
	if err != nil {
		t.Error(table)
	}
	foundValue, found := table.Get(key)
	if !found {
		t.Error("Key was not found in memtable")
	}
	if foundValue != value {
		t.Error("Wrong value in memtable")
	}
}

func TestMemtable_GetNonExistingKey(t *testing.T) {
	table := NewMemTable(memTableSize)
	_, found := table.Get([]byte("Non existing key"))
	if found {
		t.Error("Found non existing key in memtable")
	}
}

func TestMemtable_PutTomb(t *testing.T) {
	table := NewMemTable(memTableSize)
	err := table.Put([]byte(tombstone), &ValueMeta{offset: rand.Uint32(), length: rand.Uint32()})
	if err == nil {
		t.Error("Should not allow to save a tomb")
	}
}
