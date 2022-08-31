package wiskey

import "testing"

func TestRandStringBytes(t *testing.T) {
	first := RandStringBytes(10)
	second := RandStringBytes(10)
	if first == second {
		t.Error("Should not be the same")
	}
}
