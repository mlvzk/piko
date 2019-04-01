package main

import "testing"

func TestFormat(t *testing.T) {
	result := format("%[unknown]-ab%[id]c%[name]d%[id]", map[string]string{"id": "123", "name": "test"})
	if result != "unknown-ab123ctestd123" {
		t.Errorf("Format error, got: %s, expected: %s\n", result, "unknown-ab123ctestd123")
	}
}
