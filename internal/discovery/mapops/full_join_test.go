package mapops

import (
	"reflect"
	"testing"
)

func TestFullJoin(t *testing.T) {
	type test struct {
		left  map[string]int
		right map[string]int
		want  []struct{ Left, Right *string }
	}

	tests := []test{
		{
			left:  map[string]int{"a": 1, "b": 2},
			right: map[string]int{"b": 3, "c": 4},
			want: []struct{ Left, Right *string }{
				{Left: stringPtr("a"), Right: nil},
				{Left: stringPtr("b"), Right: stringPtr("b")},
				{Left: nil, Right: stringPtr("c")},
			},
		},
		{
			left:  map[string]int{},
			right: map[string]int{"c": 4},
			want: []struct{ Left, Right *string }{
				{Left: nil, Right: stringPtr("c")},
			},
		},
		{
			left:  map[string]int{"a": 1},
			right: map[string]int{},
			want: []struct{ Left, Right *string }{
				{Left: stringPtr("a"), Right: nil},
			},
		},
		{
			left:  map[string]int{},
			right: map[string]int{},
			want:  []struct{ Left, Right *string }{},
		},
	}

	for i, tt := range tests {
		got := FullJoin(tt.left, tt.right)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Test %d: got %v, want %v", i, got, tt.want)
		}
	}
}

func stringPtr(s string) *string {
	return &s
}
