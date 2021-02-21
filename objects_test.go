package main

import "testing"

func TestOverlaps(t *testing.T) {
	object := NewObject("assets/asteroid.png")
	if object == nil {
		t.Logf("could't even make an object...")
	}
	// TODO: write real collision test cases here, the above is just to stop the
	// compiler complaining about unused vars etc
}
