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

package testutil

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func CacheHttpRequest(t *testing.T, base string, update bool) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		golden := filepath.Join("testdata", strings.Replace(t.Name(), "/", "_", -1)+"-resp.golden")

		if update {
			req, err := http.NewRequest("GET", base+r.URL.Path+"?"+r.URL.RawQuery, nil)
			if err != nil {
				t.Fatalf("Error creating new request: %v", err)
			}
			req.Header = r.Header
			req.Header.Del("accept-encoding")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Error sending http request: %v", err)
			}
			defer resp.Body.Close()

			file, err := os.Create(golden)
			if err != nil {
				t.Fatalf("Error creating golden file: %v", err)
			}

			io.Copy(file, resp.Body)
			file.Close()
		}

		goldenFile, err := os.Open(golden)
		if err != nil {
			t.Fatalf("Couldn't open the golden file: %v", err)
		}
		io.Copy(w, goldenFile)
	}))

	return ts
}
