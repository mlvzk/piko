package soundcloud

import (
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mlvzk/shovel-go/service"
)

const baseApiURL = "https://api.soundcloud.com"

var update = flag.Bool("update", false, "update .golden files")

func TestIsValidTarget(t *testing.T) {
	tests := map[string]bool{
		"https://soundcloud.com/musicpromouser/mac-miller-ok-ft-tyler-the-creator": true,
		"https://soundcloud.com/":              true,
		"https://soundcloud.com/search?q=test": true,
		"https://soundcloud.com/fadermedia":    true,
		"https://instagram.com/":               false,
	}

	for target, expected := range tests {
		if (Soundcloud{}).IsValidTarget(target) != expected {
			t.Errorf("Invalid result, target: %v, expected: %v", target, expected)
		}
	}
}

func TestIteratorNext(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		golden := filepath.Join("testdata", t.Name()+"-resp.golden")

		if *update {
			resp, err := http.Get(baseApiURL + r.URL.Path + "?" + r.URL.RawQuery)
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

	iterator := SoundcloudIterator{
		url:        "https://soundcloud.com/ishaan-bhagwakar/oldie-ofwgkta",
		baseApiURL: ts.URL,
		clientID:   "a3e059563d7fd3372b49b37f00a00bcf",
	}

	items, err := iterator.Next()
	if err != nil {
		t.Fatalf("iterator.Next() error: %v", err)
	}

	for _, item := range items {
		item.Meta["playCount"] = "ignore"
		item.Meta["_downloadURL"] = "ignore"
	}

	expected := []service.Item{
		service.Item{
			Meta: map[string]string{
				"id":           "224754696",
				"title":        "Oldie - OFWGKTA",
				"username":     "Ishaan Bhagwakar",
				"createdAt":    "2015/09/20 17:32:13 +0000",
				"duration":     "636453",
				"playCount":    "ignore",
				"ext":          "mp3",
				"_downloadURL": "ignore",
			},
			DefaultName: "%[title].%[ext]",
		},
	}

	if diff := pretty.Compare(items, expected); diff != "" {
		t.Errorf("%s diff:\n%s", t.Name(), diff)
	}
}
