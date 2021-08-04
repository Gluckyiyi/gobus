package gobus

import "testing"

func TestNew(t *testing.T) {
	bus := New()
	if bus == nil {
		t.Log()
		t.Fail()
	}
}
