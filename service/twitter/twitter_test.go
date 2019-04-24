package twitter

import (
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mlvzk/piko/service"
)

const base = "https://twitter.com"

var update = flag.Bool("update", false, "update .golden files")

func TestIsValidTarget(t *testing.T) {
	tests := map[string]bool{
		"https://twitter.com/golang":                            true,
		"https://twitter.com/golang/status/1116531752602951681": true,
		"twitter.com/golang/status/1116531752602951681":         true,
		"https://soundcloud.com/":                               false,
	}

	for target, expected := range tests {
		if (Twitter{}).IsValidTarget(target) != expected {
			t.Errorf("Invalid result, target: %v, expected: %v", target, expected)
		}
	}
}

func TestIteratorNext(t *testing.T) {
	path := "/golang/status/1106303553474301955"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		golden := filepath.Join("testdata", t.Name()+"-resp.golden")

		if *update {
			resp, err := http.Get(base + path)
			if err != nil {
				t.Fatalf("Error updating golden file: %v", err)
			}
			defer resp.Body.Close()

			file, err := os.Create(golden)
			if err != nil {
				t.Fatalf("Error creating golden file")
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
	defer ts.Close()

	iterator := TwitterIterator{
		url: ts.URL + path,
	}

	items, err := iterator.Next()
	if err != nil {
		t.Fatalf("iterator.Next() error: %v", err)
	}

	if len(items) < 1 {
		t.Fatalf("Items array is empty")
	}

	if !strings.Contains(items[0].Meta["downloadURL"], "pbs.twimg.com") {
		t.Fatalf("Incorrect downloadURL")
	}
	items[0].Meta["downloadURL"] = "ignore"

	expected := []service.Item{
		service.Item{
			Meta: map[string]string{
				"downloadURL": "ignore",
				"index":       "0",
				"id":          "1106303553474301955",
				"author":      "golang",
				"description": "🎉 Go 1.12.1 and 1.11.6 are released!\n\n🗣 Announcement: https://t.co/PAttJybffj\n\nHappy Pi day! 🥧\n\n#golang",
				"type":        "image",
				"ext":         "jpg",
			},
			DefaultName: "%[author]-%[id]-%[index].%[ext]",
		},
	}

	if diff := pretty.Compare(items, expected); diff != "" {
		t.Errorf("%s diff:\n%s", t.Name(), diff)
	}
}
