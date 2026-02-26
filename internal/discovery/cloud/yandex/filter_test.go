package yandex

import (
	"testing"
)

//go:fix inline
func stringPtr(s string) *string {
	return new(s)
}

func TestMatchName(t *testing.T) {
	filter := NewFilter(".*test.*", nil, new("exclude"), nil)

	testCases := []struct {
		name    string
		want    bool
		message string
	}{
		{"test", true, "should match simple test"},
		{"mytest", true, "should match mytest"},
		{"exclude", false, "should not match exclude"},
		{"another exclude", false, "should not match another exclude"},
	}

	for _, tc := range testCases {
		got := filter.MatchName(tc.name)
		if got != tc.want {
			t.Errorf("MatchName(%q) = %v; want %v. %s", tc.name, got, tc.want, tc.message)
		}
	}
}

func TestMatchDb(t *testing.T) {
	filter := NewFilter(".*", new(".*db.*"), nil, new("exclude"))

	testCases := []struct {
		name    string
		want    bool
		message string
	}{
		{"mydb", true, "should match mydb"},
		{"testdb", true, "should match testdb"},
		{"exclude", false, "should not match exclude"},
	}

	for _, tc := range testCases {
		got := filter.MatchDb(tc.name)
		if got != tc.want {
			t.Errorf("MatchDb(%q) = %v; want %v. %s", tc.name, got, tc.want, tc.message)
		}
	}
}

func TestNewFilter(t *testing.T) {
	//Test for nil values
	filter := NewFilter(".*", nil, nil, nil)
	if filter.dbRegexp != nil {
		t.Errorf("dbRegexp should be nil")
	}
	if filter.excludeDbRegexp != nil {
		t.Errorf("excludeDbRegexp should be nil")
	}
	if filter.excludeNameRegexp != nil {
		t.Errorf("excludeNameRegexp should be nil")
	}

}
