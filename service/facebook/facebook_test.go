package facebook

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

const base = "https://facebook.com"

var update = flag.Bool("update", false, "update .golden files")

func TestIsValidTarget(t *testing.T) {
	tests := map[string]bool{
		"https://www.facebook.com/Shiba.Zero.Mika/videos/414355892680582": true,
		"https://www.facebook.com/Shiba.Zero.Mika":                        true,
		"facebook.com/Shiba.Zero.Mika/videos/414355892680582":             true,
		"https://twitter.com/":                                            false,
	}

	for target, expected := range tests {
		if (Facebook{}).IsValidTarget(target) != expected {
			t.Errorf("Invalid result, target: %v, expected: %v", target, expected)
		}
	}
}
func TestIteratorNext(t *testing.T) {
	path := "/Shiba.Zero.Mika/videos/414355892680582"

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

	iterator := FacebookIterator{
		url: ts.URL + path,
	}

	items, err := iterator.Next()
	if err != nil {
		t.Fatalf("iterator.Next() error: %v", err)
	}

	if len(items) < 1 {
		t.Fatalf("Items array is empty")
	}

	for _, item := range items {
		if !strings.Contains(item.Meta["downloadURL"], "fbcdn.net") {
			t.Fatalf("Incorrect downloadURL: %s", item.Meta["downloadURL"])
		}
		item.Meta["downloadURL"] = "ignore"
	}

	expected := []service.Item{
		service.Item{
			Meta: map[string]string{
				"id":          "8057020672024510464",
				"author":      "Shiba Inu Zero.Mika",
				"description": "早晨啊🌼今早傻波在睡夢中又滾了下床😅之後起身扮作若無其事地再上床睡😂\n#柴犬 #shiba #zeromika #shibazeromika",
				"ext":         "jpg",
				"type":        "image",
				"downloadURL": "ignore",
			},
			DefaultName: "%[author]-%[id].%[ext]",
		},
		service.Item{
			Meta: map[string]string{
				"id":          "2750275577282508998",
				"author":      "Shiba Inu Zero.Mika",
				"description": "早晨啊🌼今早傻波在睡夢中又滾了下床😅之後起身扮作若無其事地再上床睡😂\n#柴犬 #shiba #zeromika #shibazeromika",
				"ext":         "mp4",
				"type":        "video",
				"downloadURL": "ignore",
			},
			DefaultName: "%[author]-%[id].%[ext]",
		},
	}

	if diff := pretty.Compare(items, expected); diff != "" {
		t.Errorf("%s diff:\n%s", t.Name(), diff)
	}
}
