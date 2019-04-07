package main

import "testing"

func TestFormat(t *testing.T) {
	cases := []struct {
		name      string
		formatStr string
		meta      map[string]string
		expected  string
	}{
		{"empty meta", "%[unknown].png", map[string]string{}, "unknown.png"},
		{"one in meta", "ab%[id].jpg", map[string]string{"id": "456"}, "ab456.jpg"},
		{"one in meta, double usage", "%[id].%[id].jpg", map[string]string{"id": "456"}, "456.456.jpg"},
		{"two in meta", "ab%[id]c%[name]d.jpg", map[string]string{
			"id":   "123",
			"name": "test",
		}, "ab123ctestd.jpg"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := format(c.formatStr, c.meta)
			if result != c.expected {
				t.Errorf("Format error, got: %s, expected: %s\n", result, c.expected)
			}
		})
	}
}
