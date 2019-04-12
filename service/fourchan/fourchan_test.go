package fourchan

import (
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mlvzk/piko/service"
)

const base = "https://boards.4channel.org"

var update = flag.Bool("update", false, "update .golden files")

func TestFourchanIsValidTarget(t *testing.T) {
	tests := map[string]bool{
		"https://boards.4channel.org/g/": true,
		"boards.4channel.org/g/":         true,
		"https://boards.4channel.org/g/thread/70377765/hpg-esg-headphone-general": true,
		"https://boards.4chan.org/pol/":                                           true,
		"https://imgur.com/":                                                      false,
	}

	for target, expected := range tests {
		if (Fourchan{}).IsValidTarget(target) != expected {
			t.Errorf("Invalid result, target: %v, expected: %v", target, expected)
		}
	}
}

func TestIteratorNext(t *testing.T) {
	path := "/adv/thread/20765545/i-want-to-be-the-very-best-like-no-one-ever-was"

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

	iterator := FourchanIterator{
		url: ts.URL + path,
	}

	items, err := iterator.Next()
	if err != nil {
		t.Fatalf("iterator.Next() error: %v", err)
	}

	expected := []service.Item{
		service.Item{
			Meta: map[string]string{
				"title":        "1306532808724.png",
				"imgURL":       "https://i.4cdn.org/adv/1554743536847.png",
				"id":           "1554743536847",
				"ext":          "png",
				"thumbnailURL": "https://i.4cdn.org/adv/1554743536847s.jpg",
			},
			DefaultName: "%[title]",
			AvailableOptions: map[string][]string{
				"thumbnail": []string{"yes", "no"},
			},
			DefaultOptions: map[string]string{
				"thumbnail": "no",
			},
		},
		service.Item{
			Meta: map[string]string{
				"title":        "1306532948201.png",
				"imgURL":       "https://i.4cdn.org/adv/1554745883162.png",
				"id":           "1554745883162",
				"ext":          "png",
				"thumbnailURL": "https://i.4cdn.org/adv/1554745883162s.jpg",
			},
			DefaultName: "%[title]",
			AvailableOptions: map[string][]string{
				"thumbnail": []string{"yes", "no"},
			},
			DefaultOptions: map[string]string{
				"thumbnail": "no",
			},
		},
	}

	if diff := pretty.Compare(items, expected); diff != "" {
		t.Errorf("%s diff:\n%s", t.Name(), diff)
	}
}
