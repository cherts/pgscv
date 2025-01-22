// Package mapops implement join ops
package mapops

// FullJoin implement full join of maps by keys. Keys should be comparable.
func FullJoin[K comparable, V1 any, V2 any](left map[K]V1, right map[K]V2) []struct{ Left, Right *K } {
	result := make([]struct{ Left, Right *K }, 0, len(left)+len(right))
	for l := range left {
		if _, ok := right[l]; !ok {
			result = append(result, struct{ Left, Right *K }{Left: &l, Right: nil})
		} else {
			result = append(result, struct{ Left, Right *K }{Left: &l, Right: &l})
		}
	}
	for r := range right {
		if _, ok := left[r]; !ok {
			result = append(result, struct{ Left, Right *K }{Left: nil, Right: &r})
		}
	}
	return result
}
