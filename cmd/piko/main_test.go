// Copyright 2019 mlvzk
// This file is part of the piko library.
//
// The piko library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The piko library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the piko library. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"testing"
)

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
