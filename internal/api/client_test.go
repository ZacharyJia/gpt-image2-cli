package api

import "testing"

func TestResolveSize(t *testing.T) {
	c, err := New("dummy-key", "")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		in   string
		want string
	}{
		{"1k", "1024x1024"},
		{"2k", "2048x2048"},
		{"portrait", "1024x1536"},
		{"landscape", "1536x1024"},
		{"1024x1024", "1024x1024"},
		{"", ""},
	}
	for _, tc := range cases {
		got := c.resolveSize(tc.in)
		if got != tc.want {
			t.Errorf("resolveSize(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestStartsWithIgnoreCase(t *testing.T) {
	if !startsWithIgnoreCase("GPT-Image-2-foo", "gpt-image-2") {
		t.Error("expected true")
	}
	if startsWithIgnoreCase("gpt-image-1", "gpt-image-2") {
		t.Error("expected false")
	}
}
