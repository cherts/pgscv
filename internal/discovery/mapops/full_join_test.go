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
		if !equalJoinResults(got, tt.want) {
			t.Errorf("Test %d: got %v, want %v", i, got, tt.want)
		}
	}
}

func equalJoinResults(a, b []struct{ Left, Right *string }) bool {
	if len(a) != len(b) {
		return false
	}

	aMap := make(map[string]*string)
	bMap := make(map[string]*string)

	for _, item := range a {
		key := ""
		if item.Left != nil {
			key += *item.Left
		}
		if item.Right != nil {
			key += *item.Right
		}
		aMap[key] = item.Left
	}

	for _, item := range b {
		key := ""
		if item.Left != nil {
			key += *item.Left
		}
		if item.Right != nil {
			key += *item.Right
		}
		bMap[key] = item.Left
	}

	return reflect.DeepEqual(aMap, bMap)
}

func stringPtr(s string) *string {
	return &s
}
