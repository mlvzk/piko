package instagram

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
	"github.com/mlvzk/shovel-go/service"
)

const base = "https://www.instagram.com"

var update = flag.Bool("update", false, "update .golden files")

func TestIsValidTarget(t *testing.T) {
	tests := map[string]bool{
		"https://www.instagram.com/explore/tags/cat/": true,
		"instagram.com/explore/tags/cat/":             true,
		"https://www.instagram.com/p/Bv3X1rVBWm5/":    true,
		"https://www.instagram.com/newding2/?hl=en":   true,
		"https://youtube.com/":                        false,
	}

	for target, expected := range tests {
		if (Instagram{}).IsValidTarget(target) != expected {
			t.Errorf("Invalid result, target: %v, expected: %v", target, expected)
		}
	}
}

func TestIteratorNext(t *testing.T) {
	path := "/p/BsOGulcndj-/"

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

	iterator := InstagramIterator{
		url: ts.URL + path,
	}

	items, err := iterator.Next()
	if err != nil {
		t.Fatalf("iterator.Next() error: %v", err)
	}

	if len(items) < 1 {
		t.Fatalf("Items array is empty")
	}

	correctURL := strings.Contains(items[0].Meta["imgURL"], "cdninstagram.com/vp/cd672a3988fc98ee1b493914193c8545/5D4F2BB4/t51.2885-15/e35/47692668_1958135090974774_6762833792332802352_n.jpg")
	if !correctURL {
		t.Fatalf("Incorrect imgURL")
	}
	items[0].Meta["imgURL"] = "ignore"

	expected := []service.Item{
		service.Item{
			Meta: map[string]string{
				"imgURL": "ignore",
				"caption": `Let’s set a world record together and get the most liked post on Instagram. Beating the current world record held by Kylie Jenner (18 million)! We got this 🙌

#LikeTheEgg #EggSoldiers #EggGang`,
				"author": "world_record_egg",
				"date":   "2019-01-04T17:05:45",
				"id":     "BsOGulcndj-",
			},
			DefaultName: "%[title].jpg",
		},
	}

	if diff := pretty.Compare(items, expected); diff != "" {
		t.Errorf("%s diff:\n%s", t.Name(), diff)
	}
}
