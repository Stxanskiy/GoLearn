package repository

import (
	"slices"
	"testing"
)

func TestMatchByKey(t *testing.T) {
	tests := []struct {
		name               string
		existing, incoming []string
		pairs, unused      []int
	}{
		{"same order", []string{"a", "b"}, []string{"a", "b"}, []int{0, 1}, nil},
		{"reordered and inserted", []string{"a", "b", "c"}, []string{"c", "new", "a"}, []int{2, -1, 0}, []int{1}},
		{"duplicates pair in order", []string{"x", "x"}, []string{"x"}, []int{0}, []int{1}},
		{"all removed", []string{"a"}, nil, []int{}, []int{0}},
		{"all new", nil, []string{"a"}, []int{-1}, nil},
	}
	for _, tt := range tests {
		pairs, unused := matchByKey(tt.existing, tt.incoming)
		if !slices.Equal(pairs, tt.pairs) || !slices.Equal(unused, tt.unused) {
			t.Errorf("%s: pairs %v unused %v, want %v %v", tt.name, pairs, unused, tt.pairs, tt.unused)
		}
	}
}
