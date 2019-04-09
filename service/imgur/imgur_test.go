package imgur

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

const base = "https://imgur.com"

var update = flag.Bool("update", false, "update .golden files")

func TestIsValidTarget(t *testing.T) {
	tests := map[string]bool{
		"https://imgur.com/gallery/kgIfZrm":     true,
		"imgur.com/gallery/kgIfZrm":             true,
		"https://imgur.com/t/article13/y2Vp0nZ": true,
		"https://youtube.com/":                  false,
	}

	for target, expected := range tests {
		if (Imgur{}).IsValidTarget(target) != expected {
			t.Errorf("Invalid result, target: %v, expected: %v", target, expected)
		}
	}
}

func TestIteratorNext(t *testing.T) {
	path := "/t/article13/EfY6CxU"

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

	iterator := ImgurIterator{
		url: ts.URL + path,
	}

	items, err := iterator.Next()
	if err != nil {
		t.Fatalf("iterator.Next() error: %v", err)
	}

	expected := []service.Item{
		service.Item{
			Meta: map[string]string{
				"id":         "o2nusiZ",
				"ext":        "png",
				"itemType":   "http://schema.org/ImageObject",
				"albumTitle": "Some advice for those of you in the EU",
			},
			DefaultName: "%[id].%[ext]",
		},
	}

	if diff := pretty.Compare(items, expected); diff != "" {
		t.Errorf("%s diff:\n%s", t.Name(), diff)
	}
}
