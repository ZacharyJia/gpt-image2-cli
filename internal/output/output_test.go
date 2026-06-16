package output

import "testing"

func TestSlugify(t *testing.T) {
	cases := []struct {
		in      string
		maxLen  int
		want    string
	}{
		{"Hello World", 30, "hello-world"},
		{"  Multiple   Spaces  ", 30, "multiple-spaces"},
		{"Special!@#Characters", 30, "specialcharacters"},
		{"", 30, "image"},
		{"A very long prompt that exceeds the maximum length", 10, "a-very-lon"},
	}
	for _, c := range cases {
		got := Slugify(c.in, c.maxLen)
		if got != c.want {
			t.Errorf("Slugify(%q, %d) = %q, want %q", c.in, c.maxLen, got, c.want)
		}
	}
}
