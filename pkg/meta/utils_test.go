package meta

import (
	"testing"
)

func TestGetWidth(t *testing.T) {
	// Happy path
	strs := []string{"hello", "world", "nyaago"}
	if getWidth(strs) != 6 {
		t.Errorf("Expected width 6, got %d", getWidth(strs))
	}

	// Empty slice
	strs = []string{}
	if getWidth(strs) != 0 {
		t.Errorf("Expected width 0, got %d", getWidth(strs))
	}

	// Single element
	strs = []string{"single"}
	if getWidth(strs) != 6 {
		t.Errorf("Expected width 6, got %d", getWidth(strs))
	}

	// All elements same length
	strs = []string{"test", "code", "unit"}
	if getWidth(strs) != 4 {
		t.Errorf("Expected width 4, got %d", getWidth(strs))
	}

	// Different lengths
	strs = []string{"short", "longer", "longest"}
	if getWidth(strs) != 7 {
		t.Errorf("Expected width 7, got %d", getWidth(strs))
	}

	// Including empty string
	strs = []string{"", "filler", "longeststring"}
	if getWidth(strs) != 13 {
		t.Errorf("Expected width 13, got %d", getWidth(strs))
	}

	// Edge case with one very long string
	strs = []string{"averyveryveryveryverylongstring"}
	if getWidth(strs) != 31 {
		t.Errorf("Expected width 31, got %d", getWidth(strs))
	}
}
